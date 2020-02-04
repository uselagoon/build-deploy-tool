#!/bin/bash
set -x
set -eo pipefail
set -o noglob

REGISTRY=172.17.0.1:8084 # This points to Harbor's nginx pod when running via docker-compose, which handles access to Harbor's registry
NAMESPACE=$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace)
REGISTRY_REPOSITORY=$NAMESPACE
LAGOON_VERSION=$(cat /lagoon/version)

if [ ! -z "$LAGOON_PROJECT_VARIABLES" ]; then
  INTERNAL_REGISTRY_URL=$(echo $LAGOON_PROJECT_VARIABLES | jq -r '.[] | select(.scope == "internal_container_registry") | .url')
  INTERNAL_REGISTRY_USERNAME=$(echo $LAGOON_PROJECT_VARIABLES | jq -r '.[] | select(.scope == "internal_container_registry") | .username')
  INTERNAL_REGISTRY_PASSWORD=$(echo $LAGOON_PROJECT_VARIABLES | jq -r '.[] | select(.scope == "internal_container_registry") | .password')
fi

if [ "$CI" == "true" ]; then
  CI_OVERRIDE_IMAGE_REPO=${REGISTRY}/lagoon
else
  CI_OVERRIDE_IMAGE_REPO=""
fi

if [ "$TYPE" == "pullrequest" ]; then
  /kubectl-build-deploy/scripts/git-checkout-pull-merge.sh "$SOURCE_REPOSITORY" "$PR_HEAD_SHA" "$PR_BASE_SHA"
else
  /kubectl-build-deploy/scripts/git-checkout-pull.sh "$SOURCE_REPOSITORY" "$GIT_REF"
fi

if [[ -n "$SUBFOLDER" ]]; then
  cd $SUBFOLDER
fi

if [ ! -f .lagoon.yml ]; then
  echo "no .lagoon.yml file found"; exit 1;
fi

INJECT_GIT_SHA=$(cat .lagoon.yml | shyaml get-value environment_variables.git_sha false)
if [ "$INJECT_GIT_SHA" == "true" ]
then
  LAGOON_GIT_SHA=`git rev-parse HEAD`
else
  LAGOON_GIT_SHA="0000000000000000000000000000000000000000"
fi

set +x

if [ ! -z ${INTERNAL_REGISTRY_URL} && ! -z ${INTERNAL_REGISTRY_USERNAME} && ! -z ${INTERNAL_REGISTRY_PASSWORD} ]; then
  docker login -u=${INTERNAL_REGISTRY_USERNAME} -p="${INTERNAL_REGISTRY_USERNAME}" ${INTERNAL_REGISTRY_URL}
  REGISTRY=$INTERNAL_REGISTRY_URL # This will handle pointing Lagoon at the correct registry for non local builds
  #REGISTRY_REPOSITORY=$NAMESPACE
  # If we go with a different naming scheme, we can inject that here?
#else
#  DOCKER_REGISTRY_TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
#  docker login -u=jenkins -p="${DOCKER_REGISTRY_TOKEN}" ${REGISTRY}
fi

DEPLOYER_TOKEN=$(cat /var/run/secrets/lagoon/deployer/token)

kubectl config set-credentials lagoon/kubernetes.default.svc --token="${DEPLOYER_TOKEN}"
kubectl config set-cluster kubernetes.default.svc --insecure-skip-tls-verify=true --server=https://kubernetes.default.svc
kubectl config set-context default/lagoon/kubernetes.default.svc --user=lagoon/kubernetes.default.svc --namespace="${NAMESPACE}" --cluster=kubernetes.default.svc
kubectl config use-context default/lagoon/kubernetes.default.svc

set -x

ADDITIONAL_YAMLS=($(cat .lagoon.yml | shyaml keys additional-yaml || echo ""))

for ADDITIONAL_YAML in "${ADDITIONAL_YAMLS[@]}"
do
  ADDITIONAL_YAML_PATH=$(cat .lagoon.yml | shyaml get-value additional-yaml.$ADDITIONAL_YAML.path false)
  if [ $ADDITIONAL_YAML_PATH == "false" ]; then
    echo "No 'path' defined for additional yaml $ADDITIONAL_YAML"; exit 1;
  fi

  if [ ! -f $ADDITIONAL_YAML_PATH ]; then
    echo "$ADDITIONAL_YAML_PATH for additional yaml $ADDITIONAL_YAML not found"; exit 1;
  fi

  ADDITIONAL_YAML_COMMAND=$(cat .lagoon.yml | shyaml get-value additional-yaml.$ADDITIONAL_YAML.command apply)
  ADDITIONAL_YAML_IGNORE_ERROR=$(cat .lagoon.yml | shyaml get-value additional-yaml.$ADDITIONAL_YAML.ignore_error false)
  ADDITIONAL_YAML_IGNORE_ERROR="${ADDITIONAL_YAML_IGNORE_ERROR,,}" # convert to lowercase, as shyaml returns "True" if the yaml is set to "true"
  . /kubectl-build-deploy/scripts/exec-additional-yaml.sh
done

.  /kubectl-build-deploy/build-deploy-docker-compose.sh
