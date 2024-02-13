#!/bin/bash
set -x
set -eo pipefail
set -o noglob

set +x # reduce noise in build logs
# print out the build-deploy-tool version information
echo "##############################################"
build-deploy-tool version
echo "##############################################"
set -x

REGISTRY=$REGISTRY
NAMESPACE=$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace)
LAGOON_VERSION=$(cat /lagoon/version)

set +x # reduce noise in build logs
if [ ! -z "$LAGOON_PROJECT_VARIABLES" ]; then
  INTERNAL_REGISTRY_URL=$(jq --argjson data "$LAGOON_PROJECT_VARIABLES" -n -r '$data | .[] | select(.scope == "internal_container_registry") | select(.name == "INTERNAL_REGISTRY_URL") | .value' | sed -e 's#^http://##' | sed -e 's#^https://##')
  INTERNAL_REGISTRY_USERNAME=$(jq --argjson data "$LAGOON_PROJECT_VARIABLES" -n -r '$data | .[] | select(.scope == "internal_container_registry") | select(.name == "INTERNAL_REGISTRY_USERNAME") | .value')
  INTERNAL_REGISTRY_PASSWORD=$(jq --argjson data "$LAGOON_PROJECT_VARIABLES" -n -r '$data | .[] | select(.scope == "internal_container_registry") | select(.name == "INTERNAL_REGISTRY_PASSWORD") | .value')
fi
set -x

echo -e "##############################################\nBEGIN Checkout Repository\n##############################################"
if [ "$BUILD_TYPE" == "pullrequest" ]; then
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

echo -e "##############################################\nBEGIN Kubernetes and Container Registry Setup\n##############################################"
sleep 0.5s

set +x # reduce noise in build logs
if [[ -f "/var/run/secrets/kubernetes.io/serviceaccount/token" ]]; then
  DEPLOYER_TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
else
  if [[ -f "/var/run/secrets/lagoon/deployer/token" ]]; then
    DEPLOYER_TOKEN=$(cat /var/run/secrets/lagoon/deployer/token)
  fi
fi
if [ -z ${DEPLOYER_TOKEN} ]; then
  echo "No deployer token found"; exit 1;
fi

kubectl config set-credentials lagoon/kubernetes.default.svc --token="${DEPLOYER_TOKEN}"
kubectl config set-cluster kubernetes.default.svc --server=https://kubernetes.default.svc --certificate-authority=/run/secrets/kubernetes.io/serviceaccount/ca.crt
kubectl config set-context default/lagoon/kubernetes.default.svc --user=lagoon/kubernetes.default.svc --namespace="${NAMESPACE}" --cluster=kubernetes.default.svc
kubectl config use-context default/lagoon/kubernetes.default.svc

# log in to the provided registry if details are provided
if [ ! -z ${INTERNAL_REGISTRY_URL} ] ; then
  echo "Creating secret for internal registry access"
  if [ ! -z ${INTERNAL_REGISTRY_USERNAME} ] && [ ! -z ${INTERNAL_REGISTRY_PASSWORD} ] ; then
    echo "docker login -u '${INTERNAL_REGISTRY_USERNAME}' -p '${INTERNAL_REGISTRY_PASSWORD}' ${INTERNAL_REGISTRY_URL}" | /bin/bash
    # create lagoon-internal-registry-secret if it does not exist yet 
    # TODO: remove this, the secret is created by the remote-controller, builds only need to log in to it now
    # if ! kubectl -n ${NAMESPACE} get secret lagoon-internal-registry-secret &> /dev/null; then
    #   kubectl create secret docker-registry lagoon-internal-registry-secret --docker-server=${INTERNAL_REGISTRY_URL} --docker-username=${INTERNAL_REGISTRY_USERNAME} --docker-password=${INTERNAL_REGISTRY_PASSWORD} --dry-run -o yaml | kubectl apply -f -
    # fi
    REGISTRY=$INTERNAL_REGISTRY_URL # This will handle pointing Lagoon at the correct registry for non local builds
    echo "Set internal registry secrets for token ${INTERNAL_REGISTRY_USERNAME} in ${REGISTRY}"
  else
    if [ ! $INTERNAL_REGISTRY_USERNAME ]; then
      echo "No token created for registry ${INTERNAL_REGISTRY_URL}"; exit 1;
    fi
    if [ ! $INTERNAL_REGISTRY_PASSWORD ]; then
      echo "No password retrieved for token ${INTERNAL_REGISTRY_USERNAME} in registry ${INTERNAL_REGISTRY_URL}"; exit 1;
    fi
  fi
fi

echo -e "\n\n##############################################\nStart Build Process\n##############################################"
set -x
.  /kubectl-build-deploy/build-deploy-docker-compose.sh
