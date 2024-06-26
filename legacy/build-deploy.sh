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
REGISTRY_REPOSITORY=$NAMESPACE
LAGOON_VERSION=$(cat /lagoon/version)

set +x # reduce noise in build logs
if [ ! -z "$LAGOON_PROJECT_VARIABLES" ]; then
  INTERNAL_REGISTRY_URL=$(jq --argjson data "$LAGOON_PROJECT_VARIABLES" -n -r '$data | .[] | select(.scope == "internal_container_registry") | select(.name == "INTERNAL_REGISTRY_URL") | .value' | sed -e 's#^http://##' | sed -e 's#^https://##')
  INTERNAL_REGISTRY_USERNAME=$(jq --argjson data "$LAGOON_PROJECT_VARIABLES" -n -r '$data | .[] | select(.scope == "internal_container_registry") | select(.name == "INTERNAL_REGISTRY_USERNAME") | .value')
  INTERNAL_REGISTRY_PASSWORD=$(jq --argjson data "$LAGOON_PROJECT_VARIABLES" -n -r '$data | .[] | select(.scope == "internal_container_registry") | select(.name == "INTERNAL_REGISTRY_PASSWORD") | .value')
fi
set -x

if [ "$CI" == "true" ]; then
  CI_OVERRIDE_IMAGE_REPO=172.17.0.1:5000/lagoon
else
  CI_OVERRIDE_IMAGE_REPO=""
fi

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

INJECT_GIT_SHA=$(cat .lagoon.yml | shyaml get-value environment_variables.git_sha false)
if [ "$INJECT_GIT_SHA" == "true" ]
then
  LAGOON_GIT_SHA=`git rev-parse HEAD`
else
  LAGOON_GIT_SHA="0000000000000000000000000000000000000000"
fi

echo -e "##############################################\nBEGIN Kubernetes and Container Registry Setup\n##############################################"
sleep 0.5s

REGISTRY_SECRETS=()
PRIVATE_REGISTRY_COUNTER=0
PRIVATE_REGISTRY_URLS=()
PRIVATE_DOCKER_HUB_REGISTRY=0
PRIVATE_EXTERNAL_REGISTRY=0

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

if [ ! -z ${INTERNAL_REGISTRY_URL} ] ; then
  echo "Creating secret for internal registry access"
  if [ ! -z ${INTERNAL_REGISTRY_USERNAME} ] && [ ! -z ${INTERNAL_REGISTRY_PASSWORD} ] ; then
    echo "docker login -u '${INTERNAL_REGISTRY_USERNAME}' -p '${INTERNAL_REGISTRY_PASSWORD}' ${INTERNAL_REGISTRY_URL}" | /bin/bash
    # create lagoon-internal-registry-secret if it does not exist yet
    if ! kubectl -n ${NAMESPACE} get secret lagoon-internal-registry-secret &> /dev/null; then
      kubectl create secret docker-registry lagoon-internal-registry-secret --docker-server=${INTERNAL_REGISTRY_URL} --docker-username=${INTERNAL_REGISTRY_USERNAME} --docker-password=${INTERNAL_REGISTRY_PASSWORD} --dry-run -o yaml | kubectl apply -f -
    fi
    REGISTRY_SECRETS+=("lagoon-internal-registry-secret")
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

##############################################
### PRIVATE REGISTRY LOGINS
##############################################
# we want to be able to support private container registries
# grab all the container-registries that are defined in the `.lagoon.yml` file
function getRegistryUsernameFromEnvironmentVariables() {
  PRIVATE_CONTAINER_REGISTRY_USERNAME_OVERRIDE_KEY="REGISTRY_${PRIVATE_CONTAINER_REGISTRY}_USERNAME"
  PRIVATE_CONTAINER_REGISTRY_USERNAME_OVERRIDE_KEY_SAFE="REGISTRY_${PRIVATE_CONTAINER_REGISTRY_SAFE}_USERNAME"
  # check if we have an override password defined anywhere in the api using the supported `REGISTRY_${registry}_USERNAME` key
  # where registry name can be the uppercased "SAFE" version
  # ie, 
  # dockerhub, docker-hub, my-custom-registry
  # become
  # DOCKERHUB, DOCKER_HUB, MY_CUSTOM_REGISTRY
  if [ ! -z "$LAGOON_PROJECT_VARIABLES" ]; then
    TEMP_PRIVATE_REGISTRY_CREDENTIAL_USERNAME=($(echo $LAGOON_PROJECT_VARIABLES | jq -r '.[] | select(.scope == "container_registry" and .name == "'$PRIVATE_CONTAINER_REGISTRY_USERNAME_OVERRIDE_KEY'") | "\(.value)"'))
    if [ ! -z "$TEMP_PRIVATE_REGISTRY_CREDENTIAL_USERNAME" ]; then
      PRIVATE_CONTAINER_REGISTRY_CREDENTIAL_USERNAME=$TEMP_PRIVATE_REGISTRY_CREDENTIAL_USERNAME
      PRIVATE_CONTAINER_REGISTRY_USERNAME_SOURCE="Lagoon API project variable $PRIVATE_CONTAINER_REGISTRY_USERNAME_OVERRIDE_KEY"
    fi
  fi
  if [ ! -z "$LAGOON_ENVIRONMENT_VARIABLES" ]; then
    TEMP_PRIVATE_REGISTRY_CREDENTIAL_USERNAME=($(echo $LAGOON_ENVIRONMENT_VARIABLES | jq -r '.[] | select(.scope == "container_registry" and .name == "'$PRIVATE_CONTAINER_REGISTRY_USERNAME_OVERRIDE_KEY'") | "\(.value)"'))
    if [ ! -z "$TEMP_PRIVATE_REGISTRY_CREDENTIAL_USERNAME" ]; then
      PRIVATE_CONTAINER_REGISTRY_CREDENTIAL_USERNAME=$TEMP_PRIVATE_REGISTRY_CREDENTIAL_USERNAME
      PRIVATE_CONTAINER_REGISTRY_USERNAME_SOURCE="Lagoon API environment variable $PRIVATE_CONTAINER_REGISTRY_USERNAME_OVERRIDE_KEY"
    fi
  fi
  # check newer "safe" key
  if [ ! -z "$LAGOON_PROJECT_VARIABLES" ]; then
    TEMP_PRIVATE_REGISTRY_CREDENTIAL_USERNAME=($(echo $LAGOON_PROJECT_VARIABLES | jq -r '.[] | select(.scope == "container_registry" and .name == "'$PRIVATE_CONTAINER_REGISTRY_USERNAME_OVERRIDE_KEY_SAFE'") | "\(.value)"'))
    if [ ! -z "$TEMP_PRIVATE_REGISTRY_CREDENTIAL_USERNAME" ]; then
      PRIVATE_CONTAINER_REGISTRY_CREDENTIAL_USERNAME=$TEMP_PRIVATE_REGISTRY_CREDENTIAL_USERNAME
      PRIVATE_CONTAINER_REGISTRY_USERNAME_SOURCE="Lagoon API project variable $PRIVATE_CONTAINER_REGISTRY_USERNAME_OVERRIDE_KEY_SAFE"
    fi
  fi
  if [ ! -z "$LAGOON_ENVIRONMENT_VARIABLES" ]; then
    TEMP_PRIVATE_REGISTRY_CREDENTIAL_USERNAME=($(echo $LAGOON_ENVIRONMENT_VARIABLES | jq -r '.[] | select(.scope == "container_registry" and .name == "'$PRIVATE_CONTAINER_REGISTRY_USERNAME_OVERRIDE_KEY_SAFE'") | "\(.value)"'))
    if [ ! -z "$TEMP_PRIVATE_REGISTRY_CREDENTIAL_USERNAME" ]; then
      PRIVATE_CONTAINER_REGISTRY_CREDENTIAL_USERNAME=$TEMP_PRIVATE_REGISTRY_CREDENTIAL_USERNAME
      PRIVATE_CONTAINER_REGISTRY_USERNAME_SOURCE="Lagoon API environment variable $PRIVATE_CONTAINER_REGISTRY_USERNAME_OVERRIDE_KEY_SAFE"
    fi
  fi
}

function getRegistryPasswordFromEnvironmentVariables() {
  # check if we have a password defined anywhere in the api first that a user has specified using the older method
  # where the provided value in the password could also be an environment variable
  # this method we should look to deprecate at some stage to not have to support it
  # so maybe this could report a build warning in the future
  if [ ! -z "$LAGOON_PROJECT_VARIABLES" ]; then
    TEMP_PRIVATE_REGISTRY_CREDENTIAL=($(echo $LAGOON_PROJECT_VARIABLES | jq -r '.[] | select(.scope == "container_registry" and .name == "'$PRIVATE_CONTAINER_REGISTRY_PASSWORD'") | "\(.value)"'))
    if [ ! -z "$TEMP_PRIVATE_REGISTRY_CREDENTIAL" ]; then
      PRIVATE_REGISTRY_CREDENTIAL=$TEMP_PRIVATE_REGISTRY_CREDENTIAL
      PRIVATE_REGISTRY_CREDENTIAL_SOURCE="Lagoon API project variable $PRIVATE_CONTAINER_REGISTRY_PASSWORD"
    fi
  fi
  if [ ! -z "$LAGOON_ENVIRONMENT_VARIABLES" ]; then
    TEMP_PRIVATE_REGISTRY_CREDENTIAL=($(echo $LAGOON_ENVIRONMENT_VARIABLES | jq -r '.[] | select(.scope == "container_registry" and .name == "'$PRIVATE_CONTAINER_REGISTRY_PASSWORD'") | "\(.value)"'))
    if [ ! -z "$TEMP_PRIVATE_REGISTRY_CREDENTIAL" ]; then
      PRIVATE_REGISTRY_CREDENTIAL=$TEMP_PRIVATE_REGISTRY_CREDENTIAL
      PRIVATE_REGISTRY_CREDENTIAL_SOURCE="Lagoon API environment variable $PRIVATE_CONTAINER_REGISTRY_PASSWORD"
    fi
  fi

  PRIVATE_CONTAINER_REGISTRY_OVERRIDE_KEY="REGISTRY_${PRIVATE_CONTAINER_REGISTRY}_PASSWORD"
  PRIVATE_CONTAINER_REGISTRY_OVERRIDE_KEY_SAFE="REGISTRY_${PRIVATE_CONTAINER_REGISTRY_SAFE}_PASSWORD"
  # check if we have an override password defined anywhere in the api using the supported `REGISTRY_${registry}_USERNAME` key
  # where registry name can be the uppercased "SAFE" version
  # ie, 
  # dockerhub, docker-hub, my-custom-registry
  # become
  # DOCKERHUB, DOCKER_HUB, MY_CUSTOM_REGISTRY
  if [ ! -z "$LAGOON_PROJECT_VARIABLES" ]; then
    TEMP_PRIVATE_REGISTRY_CREDENTIAL=($(echo $LAGOON_PROJECT_VARIABLES | jq -r '.[] | select(.scope == "container_registry" and .name == "'$PRIVATE_CONTAINER_REGISTRY_OVERRIDE_KEY'") | "\(.value)"'))
    if [ ! -z "$TEMP_PRIVATE_REGISTRY_CREDENTIAL" ]; then
      PRIVATE_REGISTRY_CREDENTIAL=$TEMP_PRIVATE_REGISTRY_CREDENTIAL
      PRIVATE_REGISTRY_CREDENTIAL_SOURCE="Lagoon API project variable $PRIVATE_CONTAINER_REGISTRY_OVERRIDE_KEY"
    fi
  fi
  if [ ! -z "$LAGOON_ENVIRONMENT_VARIABLES" ]; then
    TEMP_PRIVATE_REGISTRY_CREDENTIAL=($(echo $LAGOON_ENVIRONMENT_VARIABLES | jq -r '.[] | select(.scope == "container_registry" and .name == "'$PRIVATE_CONTAINER_REGISTRY_OVERRIDE_KEY'") | "\(.value)"'))
    if [ ! -z "$TEMP_PRIVATE_REGISTRY_CREDENTIAL" ]; then
      PRIVATE_REGISTRY_CREDENTIAL=$TEMP_PRIVATE_REGISTRY_CREDENTIAL
      PRIVATE_REGISTRY_CREDENTIAL_SOURCE="Lagoon API environment variable $PRIVATE_CONTAINER_REGISTRY_OVERRIDE_KEY"
    fi
  fi
  # check newer "safe" key
  if [ ! -z "$LAGOON_PROJECT_VARIABLES" ]; then
    TEMP_PRIVATE_REGISTRY_CREDENTIAL=($(echo $LAGOON_PROJECT_VARIABLES | jq -r '.[] | select(.scope == "container_registry" and .name == "'$PRIVATE_CONTAINER_REGISTRY_OVERRIDE_KEY_SAFE'") | "\(.value)"'))
    if [ ! -z "$TEMP_PRIVATE_REGISTRY_CREDENTIAL" ]; then
      PRIVATE_REGISTRY_CREDENTIAL=$TEMP_PRIVATE_REGISTRY_CREDENTIAL
      PRIVATE_REGISTRY_CREDENTIAL_SOURCE="Lagoon API project variable $PRIVATE_CONTAINER_REGISTRY_OVERRIDE_KEY_SAFE"
    fi
  fi
  if [ ! -z "$LAGOON_ENVIRONMENT_VARIABLES" ]; then
    TEMP_PRIVATE_REGISTRY_CREDENTIAL=($(echo $LAGOON_ENVIRONMENT_VARIABLES | jq -r '.[] | select(.scope == "container_registry" and .name == "'$PRIVATE_CONTAINER_REGISTRY_OVERRIDE_KEY_SAFE'") | "\(.value)"'))
    if [ ! -z "$TEMP_PRIVATE_REGISTRY_CREDENTIAL" ]; then
      PRIVATE_REGISTRY_CREDENTIAL=$TEMP_PRIVATE_REGISTRY_CREDENTIAL
      PRIVATE_REGISTRY_CREDENTIAL_SOURCE="Lagoon API environment variable $PRIVATE_CONTAINER_REGISTRY_OVERRIDE_KEY_SAFE"
    fi
  fi
}

PRIVATE_CONTAINER_REGISTRIES=($(cat .lagoon.yml | shyaml keys container-registries 2> /dev/null || echo ""))
if [ ! -z $PRIVATE_CONTAINER_REGISTRIES ]; then
  echo -e "##############################################\nBEGIN Custom Container Registries Setup\n##############################################"
  sleep 0.5s
fi
for PRIVATE_CONTAINER_REGISTRY in "${PRIVATE_CONTAINER_REGISTRIES[@]}"
do
  echo "> Checking details for ${PRIVATE_CONTAINER_REGISTRY}"
  PRIVATE_CONTAINER_REGISTRY_SAFE=$(echo ${PRIVATE_CONTAINER_REGISTRY} | tr '[:lower:]' '[:upper:]' | tr '-' '_')
  # check if a url is set, if none set proceed against docker hub
  PRIVATE_CONTAINER_REGISTRY_URL=$(yq e '.container-registries.'$PRIVATE_CONTAINER_REGISTRY'.url' .lagoon.yml)
  if [ "$PRIVATE_CONTAINER_REGISTRY_URL" == "null" ]; then
    echo "No 'url' defined for registry $PRIVATE_CONTAINER_REGISTRY, will proceed against docker hub";
    PRIVATE_CONTAINER_REGISTRY_URL=""
  fi
  # check the username and passwords are defined in yaml
  PRIVATE_CONTAINER_REGISTRY_USERNAME=""
  PRIVATE_CONTAINER_REGISTRY_USERNAME=$(yq e '.container-registries.'$PRIVATE_CONTAINER_REGISTRY'.username' .lagoon.yml)
  
  PRIVATE_CONTAINER_REGISTRY_CREDENTIAL_USERNAME=""
  getRegistryUsernameFromEnvironmentVariables

  if [ "$PRIVATE_CONTAINER_REGISTRY_USERNAME" == "null" ] && [ -z $PRIVATE_CONTAINER_REGISTRY_CREDENTIAL_USERNAME ]; then
    echo "No 'username' defined for registry $PRIVATE_CONTAINER_REGISTRY"; exit 1;
  fi
  if [ -z $PRIVATE_CONTAINER_REGISTRY_CREDENTIAL_USERNAME ]; then
    #if no password defined in the lagoon api, pass the one in `.lagoon.yml` as a password
    PRIVATE_CONTAINER_REGISTRY_CREDENTIAL_USERNAME=$PRIVATE_CONTAINER_REGISTRY_USERNAME
    PRIVATE_CONTAINER_REGISTRY_USERNAME_SOURCE=".lagoon.yml"
  fi
  PRIVATE_CONTAINER_REGISTRY_PASSWORD=""
  PRIVATE_CONTAINER_REGISTRY_PASSWORD=$(yq e '.container-registries.'$PRIVATE_CONTAINER_REGISTRY'.password' .lagoon.yml)
  PRIVATE_REGISTRY_CREDENTIAL=""
  getRegistryPasswordFromEnvironmentVariables

  if [ "$PRIVATE_CONTAINER_REGISTRY_PASSWORD" == "null" ] && [ -z $PRIVATE_REGISTRY_CREDENTIAL ]; then
    echo "No 'password' defined for registry $PRIVATE_CONTAINER_REGISTRY"; exit 1;
  fi
  # if we have everything we need, we can proceed to logging in
  if [ -z $PRIVATE_REGISTRY_CREDENTIAL ]; then
    #if no password defined in the lagoon api, pass the one in `.lagoon.yml` as a password
    PRIVATE_REGISTRY_CREDENTIAL=$PRIVATE_CONTAINER_REGISTRY_PASSWORD
    PRIVATE_REGISTRY_CREDENTIAL_SOURCE=".lagoon.yml (we recommend using an environment variable, see the docs on container-registries for more information)"
  fi
  if [ -z "$PRIVATE_REGISTRY_CREDENTIAL" ]; then
    echo -e "A private container registry ${PRIVATE_CONTAINER_REGISTRY} was defined in the .lagoon.yml file, but no password could be found in either the .lagoon.yml or in the Lagoon API\n\nPlease check if the password has been set correctly."
    exit 1
  fi
  if [ ! -z $PRIVATE_CONTAINER_REGISTRY_URL ]; then
    echo "Attempting to log in to $PRIVATE_CONTAINER_REGISTRY_URL with user $PRIVATE_CONTAINER_REGISTRY_CREDENTIAL_USERNAME from $PRIVATE_CONTAINER_REGISTRY_USERNAME_SOURCE"
    echo "Using password sourced from $PRIVATE_REGISTRY_CREDENTIAL_SOURCE"
    docker login --username $PRIVATE_CONTAINER_REGISTRY_CREDENTIAL_USERNAME --password $PRIVATE_REGISTRY_CREDENTIAL $PRIVATE_CONTAINER_REGISTRY_URL
    kubectl create secret docker-registry "lagoon-private-registry-${PRIVATE_REGISTRY_COUNTER}-secret" --docker-server=$PRIVATE_CONTAINER_REGISTRY_URL --docker-username=$PRIVATE_CONTAINER_REGISTRY_CREDENTIAL_USERNAME --docker-password=$PRIVATE_REGISTRY_CREDENTIAL --dry-run -o yaml | kubectl apply -f -
    REGISTRY_SECRETS+=("lagoon-private-registry-${PRIVATE_REGISTRY_COUNTER}-secret")
    PRIVATE_REGISTRY_URLS+=($PRIVATE_CONTAINER_REGISTRY_URL)
    PRIVATE_EXTERNAL_REGISTRY=1
    let ++PRIVATE_REGISTRY_COUNTER
  else
    echo "Attempting to log in to docker hub with user $PRIVATE_CONTAINER_REGISTRY_CREDENTIAL_USERNAME from $PRIVATE_CONTAINER_REGISTRY_USERNAME_SOURCE"
    echo "Using password sourced from $PRIVATE_REGISTRY_CREDENTIAL_SOURCE"
    docker login --username $PRIVATE_CONTAINER_REGISTRY_CREDENTIAL_USERNAME --password $PRIVATE_REGISTRY_CREDENTIAL
    kubectl create secret docker-registry "lagoon-private-registry-${PRIVATE_REGISTRY_COUNTER}-secret" --docker-server="https://index.docker.io/v1/" --docker-username=$PRIVATE_CONTAINER_REGISTRY_CREDENTIAL_USERNAME --docker-password=$PRIVATE_REGISTRY_CREDENTIAL --dry-run -o yaml | kubectl apply -f -
    REGISTRY_SECRETS+=("lagoon-private-registry-${PRIVATE_REGISTRY_COUNTER}-secret")
    PRIVATE_REGISTRY_URLS+=("")
    PRIVATE_DOCKER_HUB_REGISTRY=1
    let ++PRIVATE_REGISTRY_COUNTER
  fi
done
if [ ! -z $PRIVATE_CONTAINER_REGISTRIES ]; then
  echo -e "##############################################\nEND Custom Container Registries Setup\n##############################################"
  sleep 0.5s
fi

echo -e "\n\n##############################################\nStart Build Process\n##############################################"
set -x
.  /kubectl-build-deploy/build-deploy-docker-compose.sh
