#!/bin/bash

# get the buildname from the pod, $HOSTNAME contains this in the running pod, so we can use this
# set it to something usable here
export LAGOON_BUILD_NAME=$HOSTNAME

BUILD_WARNING_COUNT=0

function cronScheduleMoreOftenThan30Minutes() {
  #takes a unexpanded cron schedule, returns 0 if it's more often that 30 minutes
  MINUTE=$(echo $1 | (read -a ARRAY; echo ${ARRAY[0]}) )
  if [[ $MINUTE =~ ^(M|H|\*)\/([0-5]?[0-9])$ ]]; then
    # Match found for M/xx, H/xx or */xx
    # Check if xx is smaller than 30, which means this cronjob runs more often than every 30 minutes.
    STEP=${BASH_REMATCH[2]}
    if [ $STEP -lt 30 ]; then
      return 0
    else
      return 1
    fi
  elif [[ $MINUTE =~ ^\*$ ]]; then
    # We are running every minute
    return 0
  else
    # all other cases are more often than 30 minutes
    return 1
  fi
}

function contains() {
    [[ $1 =~ (^|[[:space:]])$2($|[[:space:]]) ]] && return 0 || return 1
}

# featureFlag searches for feature flag variables in the following locations
# and order:
#
# 1. The cluster-force feature flag, prefixed with LAGOON_FEATURE_FLAG_FORCE_,
#    as a build pod environment variable. This is set via a flag on the
#    build-deploy controller. This overrides the other variables and allows
#    policy enforcement at the cluster level.
#
# 2. The regular feature flag, prefixed with LAGOON_FEATURE_FLAG_, in the
#    Lagoon environment global scoped env-vars. This allows policy control at
#    the environment level.
#
# 3. The regular feature flag, prefixed with LAGOON_FEATURE_FLAG_, in the
#    Lagoon project global scoped env-vars. This allows policy control at the
#    project level. Lagoon core consolidates all env-vars into the environment.
#    Project env-vars are only checked for backwards compatibility.
#
# 4. The cluster-default feature flag, prefixed with
#    LAGOON_FEATURE_FLAG_DEFAULT_, as a build pod environment variable. This is
#    set via a flag on the build-deploy controller. This allows default policy
#    to be set at the cluster level, but maintains the ability to selectively
#    override at the project or environment level.
#
# The value of the first variable found is printed to stdout. If the variable
# is not found, print an empty string. Additional arguments are ignored.
function featureFlag() {
	# check for argument
	[ "$1" ] || return

	local forceFlagVar defaultFlagVar flagVar

	# check build pod environment for the force policy first
	forceFlagVar="LAGOON_FEATURE_FLAG_FORCE_$1"
	[ "${!forceFlagVar}" ] && echo "${!forceFlagVar}" && return

	flagVar="LAGOON_FEATURE_FLAG_$1"
	# check Lagoon environment variables
	flagValue=$(jq -r '.[] | select(.scope == "global" and .name == "'"$flagVar"'") | .value' <<<"$LAGOON_ENVIRONMENT_VARIABLES")
	[ "$flagValue" ] && echo "$flagValue" && return
	# check Lagoon project variables
	flagValue=$(jq -r '.[] | select(.scope == "global" and .name == "'"$flagVar"'") | .value' <<<"$LAGOON_PROJECT_VARIABLES")
	[ "$flagValue" ] && echo "$flagValue" && return

	# fall back to the default, if set.
	defaultFlagVar="LAGOON_FEATURE_FLAG_DEFAULT_$1"
	echo "${!defaultFlagVar}"
}

# Checks for a build/runtime/global scoped env var from Lagoon API. All env vars
# are consolidated into the environment, project env-vars are only checked for
# backwards compatibility.
function apiEnvVarCheck() {
  # check for argument
  [ "$1" ] || return

  local flagVar

  flagVar="$1"
  # check Lagoon environment variables
  flagValue=$(jq -r '.[] | select(.scope == "build" or .scope == "runtime" or .scope == "global") | select(.name == "'"$flagVar"'") | .value' <<< "$LAGOON_ENVIRONMENT_VARIABLES")
  [ "$flagValue" ] && echo "$flagValue" && return
  # check Lagoon project variables
  flagValue=$(jq -r '.[] | select(.scope == "build" or .scope == "runtime" or .scope == "global") | select(.name == "'"$flagVar"'") | .value' <<< "$LAGOON_PROJECT_VARIABLES")
  [ "$flagValue" ] && echo "$flagValue" && return

  echo "$2"
}

# Checks for a build scoped env var from Lagoon API. All env vars
# are consolidated into the environment, project env-vars are only checked for
# backwards compatibility.
function buildEnvVarCheck() {
  # check for argument
  [ "$1" ] || return

  local flagVar

  flagVar="$1"
  # check Lagoon environment variables
  flagValue=$(jq -r '.[] | select(.scope == "build") | select(.name == "'"$flagVar"'") | .value' <<< "$LAGOON_ENVIRONMENT_VARIABLES")
  [ "$flagValue" ] && echo "$flagValue" && return
  # check Lagoon project variables
  flagValue=$(jq -r '.[] | select(.scope == "build") | select(.name == "'"$flagVar"'") | .value' <<< "$LAGOON_PROJECT_VARIABLES")
  [ "$flagValue" ] && echo "$flagValue" && return

  echo "$2"
}

# Checks for a internal_container_registry scoped env var. These are set in
# lagoon-remote.
function internalContainerRegistryCheck() {
  # check for argument
  [ "$1" ] || return

  local flagVar

  flagVar="$1"
  # check Lagoon environment variables
  flagValue=$(jq -r '.[] | select(.scope == "internal_container_registry" and .name == "'"$flagVar"'") | .value' <<< "$LAGOON_ENVIRONMENT_VARIABLES")
  [ "$flagValue" ] && echo "$flagValue" && return
  # check Lagoon project variables
  flagValue=$(jq -r '.[] | select(.scope == "internal_container_registry" and .name == "'"$flagVar"'") | .value' <<< "$LAGOON_PROJECT_VARIABLES")
  [ "$flagValue" ] && echo "$flagValue" && return

  echo "$2"
}

SCC_CHECK=$(kubectl -n ${NAMESPACE} get pod ${LAGOON_BUILD_NAME} -o json | jq -r '.metadata.annotations."openshift.io/scc" // false')

function beginBuildStep() {
  [ "$1" ] || return #Buildstep start
  [ "$2" ] || return #buildstep

  echo -e "##############################################\nBEGIN ${1}\n##############################################"

  # patch the buildpod with the buildstep
  if [ "${SCC_CHECK}" == false ]; then
    kubectl patch -n ${NAMESPACE} pod ${LAGOON_BUILD_NAME} \
      -p "{\"metadata\":{\"labels\":{\"lagoon.sh/buildStep\":\"${2}\"}}}" &> /dev/null
    # tiny sleep to allow patch to complete before logs roll again
    sleep 0.5s
  fi
}

function patchBuildStep() {
  [ "$1" ] || return #total start time
  [ "$2" ] || return #step start time
  [ "$3" ] || return #previous step end time
  [ "$4" ] || return #namespace
  [ "$5" ] || return #buildstep
  [ "$6" ] || return #buildstep
  [ "$7" ] || return #has warnings
  totalStartTime=$(date -d "${1}" +%s)
  startTime=$(date -d "${2}" +%s)
  endTime=$(date -d "${3}" +%s)
  timeZone=$(date +"%Z")

  diffSeconds="$(($endTime-$startTime))"
  diffTime=$(date -d @${diffSeconds} +"%H:%M:%S" -u)

  diffTotalSeconds="$(($endTime-$totalStartTime))"
  diffTotalTime=$(date -d @${diffTotalSeconds} +"%H:%M:%S" -u)

  hasWarnings=""
  if [ "${7}" == "true" ]; then
    hasWarnings=" WithWarnings"
  fi

  echo -e "##############################################\nSTEP ${6}: Completed at ${3} (${timeZone}) Duration ${diffTime} Elapsed ${diffTotalTime}${hasWarnings}\n##############################################"
}

##############################################
### PREPARATION
##############################################

buildStartTime="$(date +"%Y-%m-%d %H:%M:%S")"
beginBuildStep "Initial Environment Collection" "collectEnvironment"

##############################################
### COLLECT INFORMATION
##############################################

# run the collector
ENVIRONMENT_DATA=$(build-deploy-tool collect environment)
# echo "$ENVIRONMENT_DATA" | jq -r '.deployments.items[]?.name'
# echo "$ENVIRONMENT_DATA" | jq -r '.cronjobs.items[]?.name'
# echo "$ENVIRONMENT_DATA" | jq -r '.ingress.items[]?.name'
# echo "$ENVIRONMENT_DATA" | jq -r '.services.items[]?.name'
# echo "$ENVIRONMENT_DATA" | jq -r '.secrets.items[]?.name'
# echo "$ENVIRONMENT_DATA" | jq -r '.pvcs.items[]?.name'
# echo "$ENVIRONMENT_DATA" | jq -r '.schedulesv1.items[]?.name'
# echo "$ENVIRONMENT_DATA" | jq -r '.schedulesv1alpha1.items[]?.name'
# echo "$ENVIRONMENT_DATA" | jq -r '.prebackuppodsv1.items[]?.name'
# echo "$ENVIRONMENT_DATA" | jq -r '.prebackuppodsv1alpha1.items[]?.name'
# echo "$ENVIRONMENT_DATA" | jq -r '.mariadbconsumers.items[]?.name'
# echo "$ENVIRONMENT_DATA" | jq -r '.mongodbconsumers.items[]?.name'
# echo "$ENVIRONMENT_DATA" | jq -r '.postgresqlconsumers.items[]?.name'

currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
patchBuildStep "${buildStartTime}" "${buildStartTime}" "${currentStepEnd}" "${NAMESPACE}" "collectEnvironment" "Initial Environment Collection" "false"
previousStepEnd=${currentStepEnd}
beginBuildStep "Initial Environment Setup" "initialSetup"
echo "STEP: Preparation started ${buildStartTime}"

# set the imagecache registry if it is provided
IMAGECACHE_REGISTRY=""
if [ ! -z "$(featureFlag IMAGECACHE_REGISTRY)" ]; then
  IMAGECACHE_REGISTRY=$(featureFlag IMAGECACHE_REGISTRY)
  # add trailing slash if it is missing
  length=${#IMAGECACHE_REGISTRY}
  last_char=${IMAGECACHE_REGISTRY:length-1:1}
  [[ $last_char != "/" ]] && IMAGECACHE_REGISTRY="$IMAGECACHE_REGISTRY/"; :
fi

set +e
currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
patchBuildStep "${buildStartTime}" "${buildStartTime}" "${currentStepEnd}" "${NAMESPACE}" "initialSetup" "Initial Environment Setup" "false"
previousStepEnd=${currentStepEnd}

# Validate `lagoon.yml` first to try detect any errors here first
beginBuildStep ".lagoon.yml Validation" "lagoonYmlValidation"
##############################################
### RUN lagoon-yml validation against the final data which may have overrides
### from .lagoon.override.yml file or LAGOON_YAML_OVERRIDE environment variable
##############################################
lyvOutput=$(bash -c 'build-deploy-tool validate lagoon-yml; exit $?' 2>&1)
lyvExit=$?

echo "Updating lagoon-yaml configmap with a pre-deploy version of the .lagoon.yml file"
if kubectl -n ${NAMESPACE} get configmap lagoon-yaml &> /dev/null; then
  # replace it
  # if the environment has already been deployed with an existing configmap that had the file in the key `.lagoon.yml`
  # just nuke the entire configmap and replace it with our new key and file
  LAGOON_YML_CM=$(kubectl -n ${NAMESPACE} get configmap lagoon-yaml -o json)
  if [ "$(echo ${LAGOON_YML_CM} | jq -r '.data.".lagoon.yml" // false')" == "false" ]; then
    # if the key doesn't exist, then just update the pre-deploy yaml only
    kubectl -n ${NAMESPACE} get configmap lagoon-yaml -o json | jq --arg add "`cat .lagoon.yml`" '.data."pre-deploy" = $add' | kubectl apply -f -
  else
    # if the key does exist, then nuke it and put the new key
    kubectl -n ${NAMESPACE} create configmap lagoon-yaml --from-file=pre-deploy=.lagoon.yml -o yaml --dry-run=client | kubectl replace -f -
  fi
 else
  # create it
  kubectl -n ${NAMESPACE} create configmap lagoon-yaml --from-file=pre-deploy=.lagoon.yml
fi

if [ "${lyvExit}" != "0" ]; then
  currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
  patchBuildStep "${buildStartTime}" "${previousStepEnd}" "${currentStepEnd}" "${NAMESPACE}" "lagoonYmlValidationError" ".lagoon.yml Validation" "false"
  previousStepEnd=${currentStepEnd}
  echo "
##############################################
Warning!
There are issues with your .lagoon.yml file that must be fixed.
Refer to the .lagoon.yml docs for the correct syntax
https://docs.lagoon.sh/using-lagoon-the-basics/lagoon-yml/
##############################################
"
  echo "${lyvOutput}"
  echo "
##############################################"
  exit 1
fi

# The attempt to valid the `docker-compose.yaml` file
beginBuildStep "Docker Compose Validation" "dockerComposeValidation"

# Load path of docker-compose that should be used
DOCKER_COMPOSE_YAML=($(cat .lagoon.yml | yq -o json | jq -r '."docker-compose-yaml"'))

DOCKER_COMPOSE_WARNING_COUNT=0
##############################################
### RUN docker compose config check against the provided docker-compose file
### use the `build-validate` built in validater to run over the provided docker-compose file
##############################################
dccOutput=$(bash -c 'build-deploy-tool validate docker-compose --docker-compose '${DOCKER_COMPOSE_YAML}'; exit $?' 2>&1)
dccExit=$?

echo "Updating docker-compose-yaml configmap with a pre-deploy version of the docker-compose.yml file"
if kubectl -n ${NAMESPACE} get configmap docker-compose-yaml &> /dev/null; then
  # replace it
  # if the environment has already been deployed with an existing configmap that had the file in the key `docker-compose.yml`
  # just nuke the entire configmap and replace it with our new key and file
  LAGOON_YML_CM=$(kubectl -n ${NAMESPACE} get configmap docker-compose-yaml -o json)
  if [ "$(echo ${LAGOON_YML_CM} | jq -r '.data."docker-compose.yml" // false')" == "false" ]; then
    # if the key doesn't exist, then just update the pre-deploy yaml only
    kubectl -n ${NAMESPACE} get configmap docker-compose-yaml -o json | jq --arg add "`cat ${DOCKER_COMPOSE_YAML}`" '.data."pre-deploy" = $add' | kubectl apply -f -
  else
    # if the key does exist, then nuke it and put the new key
    kubectl -n ${NAMESPACE} create configmap docker-compose-yaml --from-file=pre-deploy=${DOCKER_COMPOSE_YAML} -o yaml --dry-run=client | kubectl replace -f -
  fi
 else
  # create it
  kubectl -n ${NAMESPACE} create configmap docker-compose-yaml --from-file=pre-deploy=${DOCKER_COMPOSE_YAML}
fi

if [ "${dccExit}" != "0" ]; then
  currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
  patchBuildStep "${buildStartTime}" "${previousStepEnd}" "${currentStepEnd}" "${NAMESPACE}" "dockerComposeValidationError" "Docker Compose Validation" "false"
  previousStepEnd=${currentStepEnd}
  echo "
##############################################
Warning!
There are issues with your docker compose file that lagoon uses that should be fixed.
You can run docker compose config locally to check that your docker-compose file is valid.
##############################################
"
  echo ${dccOutput}
  echo "
##############################################"
  exit 1
fi

## validate the docker-compose in a way to eventually phase out forked library by displaying warnings
dccOutput=$(bash -c 'build-deploy-tool validate docker-compose --ignore-non-string-key-errors=false --ignore-missing-env-files=false --docker-compose '${DOCKER_COMPOSE_YAML}'; exit $?' 2>&1)
dccExit=$?
if [ "${dccExit}" != "0" ]; then
  ((++BUILD_WARNING_COUNT))
  ((++DOCKER_COMPOSE_WARNING_COUNT))
  echo "
##############################################
Warning!
There are issues with your docker compose file that lagoon uses that should be fixed.
You can run docker compose config locally to check that your docker-compose file is valid.
"
  if [[ "${dccOutput}" =~ "no such file or directory" ]]; then
    echo "> an env_file is defined in your docker-compose file, but no matching file found."
  fi
  if [[ "${dccOutput}" =~ "Non-string key" ]]; then
    echo "> an invalid string key was detected in your docker-compose file."
  fi
  echo ERR: ${dccOutput}
  echo ""
fi

dccOutput=$(bash -c 'build-deploy-tool validate docker-compose-with-errors --docker-compose '${DOCKER_COMPOSE_YAML}'; exit $?' 2>&1)
dccExit2=$?
if [ "${dccExit2}" != "0" ]; then
  ((++DOCKER_COMPOSE_WARNING_COUNT))
  if [ "${dccExit}" == "0" ]; then

    # this logic is to phase rollout of https://github.com/uselagoon/build-deploy-tool/pull/304
    # anything returned by this section will be a yaml error that we need to check if the feature to enable/disable errors
    # is configured, and that the environment type matches.
    # eventually this logic will be changed entirely from warnings to errors
    DOCKER_COMPOSE_VALIDATION_ERROR=false
    DOCKER_COMPOSE_VALIDATION_ERROR_VARIABLE=LAGOON_FEATURE_FLAG_DEVELOPMENT_DOCKER_COMPOSE_VALIDATION
    # this logic will make development environments return an error by default
    # adding LAGOON_FEATURE_FLAG_DEVELOPMENT_DOCKER_COMPOSE_VALIDATION=disabled can be used to disable the error and revert to a warning per project or environment
    # or add LAGOON_FEATURE_FLAG_DEFAULT_DEVELOPMENT_DOCKER_COMPOSE_VALIDATION=disabled to the remote-controller as a default to disable for a cluster
    if [[ "$(featureFlag DEVELOPMENT_DOCKER_COMPOSE_VALIDATION)" != disabled ]] && [[ "$ENVIRONMENT_TYPE" == "development" ]]; then
      DOCKER_COMPOSE_VALIDATION_ERROR=true
    fi
    # by default, production environments won't return an error unless the feature flag is enabled.
    # this allows using the feature flag to selectively apply to production environments if required
    # adding LAGOON_FEATURE_FLAG_PRODUCTION_DOCKER_COMPOSE_VALIDATION=enabled can be used to enable the error per project or environment
    # or add LAGOON_FEATURE_FLAG_DEFAULT_PRODUCTION_DOCKER_COMPOSE_VALIDATION=enabled to the remote-controller as a default to disable for a cluster
    if [[ "$(featureFlag PRODUCTION_DOCKER_COMPOSE_VALIDATION)" = enabled ]] && [[ "$ENVIRONMENT_TYPE" == "production" ]]; then
      DOCKER_COMPOSE_VALIDATION_ERROR=true
    DOCKER_COMPOSE_VALIDATION_ERROR_VARIABLE=LAGOON_FEATURE_FLAG_PRODUCTION_DOCKER_COMPOSE_VALIDATION
    fi

    ((++BUILD_WARNING_COUNT))
    echo "
##############################################"
    if [[ "$DOCKER_COMPOSE_VALIDATION_ERROR" == "true" ]]; then
      echo "Error!"
    else
      echo "Warning!"
    fi

    echo "There are issues with your docker compose file that lagoon uses that should be fixed.
You can run docker compose config locally to check that your docker-compose file is valid.
"
  fi
  echo "> There are yaml validation errors in your docker-compose file that should be corrected."
  echo ERR: ${dccOutput}
  echo ""
fi

if [[ "$DOCKER_COMPOSE_WARNING_COUNT" -gt 0 ]]; then
  echo "Read the docs for more on errors displayed here ${LAGOON_FEATURE_FLAG_DEFAULT_DOCUMENTATION_URL}/lagoon-build-errors
"
  echo "##############################################"
  currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
  patchBuildStep "${buildStartTime}" "${previousStepEnd}" "${currentStepEnd}" "${NAMESPACE}" "dockerComposeValidationWarning" "Docker Compose Validation" "true"
  previousStepEnd=${currentStepEnd}
else
  currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
  patchBuildStep "${buildStartTime}" "${previousStepEnd}" "${currentStepEnd}" "${NAMESPACE}" "dockerComposeValidation" "Docker Compose Validation" "false"
  previousStepEnd=${currentStepEnd}
fi
set -e

# Validate .lagoon.yml only, no overrides. lagoon-linter still has checks that
# aren't in build-deploy-tool.
if ! lagoon-linter; then
	echo "${LAGOON_FEATURE_FLAG_DEFAULT_DOCUMENTATION_URL}/lagoon/using-lagoon-the-basics/lagoon-yml#restrictions describes some possible reasons for this build failure."
	echo "If you require assistance to fix this error, please contact support."
	exit 1
else
	echo "lagoon-linter found no issues with the .lagoon.yml file"
fi

##################
# build deploy-tool can collect this value now from the lagoon.yml file
# this means further use of `LAGOON_GIT_SHA` can eventually be
# completely handled with build-deploy-tool wherever this value could be consumed
# this logic can then just be replaced entirely with a single export so that the build-deploy-tool
# will know what the value is, and performs the switch based on what the lagoon.yml provides
# this is retained for now until the remaining functionality that uses it is moved to the build-deploy-tool
#
#   export LAGOON_GIT_SHA=`git rev-parse HEAD`
#
INJECT_GIT_SHA=$(cat .lagoon.yml | yq -o json | jq -r '.environment_variables.git_sha // false')
if [ "$INJECT_GIT_SHA" == "true" ]
then
  # export this so the build-deploy-tool can read it
  export LAGOON_GIT_SHA=`git rev-parse HEAD`
else
  export LAGOON_GIT_SHA="0000000000000000000000000000000000000000"
fi
##################

currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
patchBuildStep "${buildStartTime}" "${previousStepEnd}" "${currentStepEnd}" "${NAMESPACE}" "lagoonYmlValidation" ".lagoon.yml Validation" "false"
previousStepEnd=${currentStepEnd}
beginBuildStep "Configure Variables" "configuringVariables"

# Load all Services that are defined
COMPOSE_SERVICES=($(cat $DOCKER_COMPOSE_YAML | yq -o json | jq -r '.services | keys_unsorted | .[]'))

##############################################
### CACHE IMAGE LIST GENERATION
##############################################

# get a list of the images in the deployments for seeing image cache if required
export LAGOON_CACHE_BUILD_ARGS=$(kubectl -n ${NAMESPACE} get deployments -o yaml -l 'lagoon.sh/service' \
  | yq -o json e '.items[].spec.template.spec.containers[].image | capture("^(?P<image>.+\/.+\/.+\/(?P<name>.+)\@.*)$")' \
  | jq -sMrc)

# Figure out which services should we handle
SERVICE_TYPES=()
IMAGES=()
NATIVE_CRONJOB_CLEANUP_ARRAY=()
declare -A MAP_DEPLOYMENT_SERVICETYPE_TO_IMAGENAME
declare -A MAP_SERVICE_TYPE_TO_COMPOSE_SERVICE
declare -A MAP_SERVICE_NAME_TO_IMAGENAME
declare -A MAP_SERVICE_NAME_TO_SERVICEBROKER_CLASS
declare -A MAP_SERVICE_NAME_TO_SERVICEBROKER_PLAN
declare -A MAP_SERVICE_NAME_TO_DBAAS_ENVIRONMENT
# this array stores the images that will need to be pulled from an external registry (private, dockerhub)
declare -A IMAGES_PULL
# this array stores the built images
declare -A IMAGES_BUILD
# this array stores the image names that will be pushed (registry/project/environment/service:tag)
declare -A IMAGES_PUSH
# this array stores the images from the source environment that will be pulled from
declare -A IMAGES_PROMOTE
# this array stores the hashes of the built images
declare -A IMAGE_HASHES
# this array stores the dbaas consumer specs
declare -A MARIADB_DBAAS_CONSUMER_SPECS
declare -A POSTGRES_DBAAS_CONSUMER_SPECS
declare -A MONGODB_DBAAS_CONSUMER_SPECS

# Allow the servicetype be overridden by the lagoon API
# This accepts colon separated values like so `SERVICE_NAME:SERVICE_TYPE_OVERRIDE`, and multiple overrides
# separated by commas
# Example 1: mariadb:mariadb-dbaas < tells any docker-compose services named mariadb to use the mariadb-dbaas service type
# Example 2: mariadb:mariadb-dbaas,nginx:nginx-persistent
TEMP_LAGOON_SERVICE_TYPES=$(apiEnvVarCheck LAGOON_SERVICE_TYPES)
if [ -n "$TEMP_LAGOON_SERVICE_TYPES" ]; then
  LAGOON_SERVICE_TYPES=$TEMP_LAGOON_SERVICE_TYPES
fi

# loop through created DBAAS templates
DBAAS=($(build-deploy-tool identify dbaas))
for COMPOSE_SERVICE in "${COMPOSE_SERVICES[@]}"
do
  # The name of the service can be overridden, if not we use the actual servicename
  SERVICE_NAME=$(cat $DOCKER_COMPOSE_YAML | yq -o json | jq -r '.services.'\"$COMPOSE_SERVICE\"'.labels."lagoon.name" // "default"')
  if [ "$SERVICE_NAME" == "default" ]; then
    SERVICE_NAME=$COMPOSE_SERVICE
  fi

  # Load the servicetype. If it's "none" we will not care about this service at all
  SERVICE_TYPE=$(cat $DOCKER_COMPOSE_YAML | yq -o json | jq -r '.services.'\"$COMPOSE_SERVICE\"'.labels."lagoon.type" // "custom"')

  # Allow the servicetype to be overriden by environment in .lagoon.yml
  ENVIRONMENT_SERVICE_TYPE_OVERRIDE=$(cat .lagoon.yml | yq -o json | jq -r '.environments.'\"${BRANCH}\"'.types.'\"$SERVICE_NAME\"' // false')
  if [ ! $ENVIRONMENT_SERVICE_TYPE_OVERRIDE == "false" ]; then
    SERVICE_TYPE=$ENVIRONMENT_SERVICE_TYPE_OVERRIDE
  fi

  if [ ! -z "$LAGOON_SERVICE_TYPES" ]; then
    IFS=',' read -ra LAGOON_SERVICE_TYPES_SPLIT <<< "$LAGOON_SERVICE_TYPES"
    for LAGOON_SERVICE_TYPE in "${LAGOON_SERVICE_TYPES_SPLIT[@]}"
    do
      IFS=':' read -ra LAGOON_SERVICE_TYPE_SPLIT <<< "$LAGOON_SERVICE_TYPE"
      if [ "${LAGOON_SERVICE_TYPE_SPLIT[0]}" == "$SERVICE_NAME" ]; then
        SERVICE_TYPE=${LAGOON_SERVICE_TYPE_SPLIT[1]}
      fi
    done
  fi

  # Previous versions of Lagoon used "python-ckandatapusher", this should be mapped to "python"
  if [[ "$SERVICE_TYPE" == "python-ckandatapusher" ]]; then
    SERVICE_TYPE="python"
  fi

  if [[ "$SERVICE_TYPE" == "opensearch" ]] || [[ "$SERVICE_TYPE" == "elasticsearch" ]]; then
    if kubectl -n ${NAMESPACE} get prebackuppods.backup.appuio.ch "${SERVICE_NAME}-prebackuppod" &> /dev/null; then
      kubectl -n ${NAMESPACE} delete prebackuppods.backup.appuio.ch "${SERVICE_NAME}-prebackuppod"
    fi
  fi

  if [ "$SERVICE_TYPE" == "none" ]; then
    continue
  fi

  # For DeploymentConfigs with multiple Services inside (like nginx-php), we allow to define the service type of within the
  # deploymentconfig via lagoon.deployment.servicetype. If this is not set we use the Compose Service Name
  DEPLOYMENT_SERVICETYPE=$(cat $DOCKER_COMPOSE_YAML | yq -o json | jq -r '.services.'\"$COMPOSE_SERVICE\"'.labels."lagoon.deployment.servicetype" // "default"')
  if [ "$DEPLOYMENT_SERVICETYPE" == "default" ]; then
    DEPLOYMENT_SERVICETYPE=$COMPOSE_SERVICE
  fi

  # The ImageName is the same as the Name of the Docker Compose ServiceName
  IMAGE_NAME=$COMPOSE_SERVICE

  for DBAAS_ENTRY in "${DBAAS[@]}"
  do
    IFS=':' read -ra DBAAS_ENTRY_SPLIT <<< "$DBAAS_ENTRY"
    DBAAS_SERVICE_NAME=${DBAAS_ENTRY_SPLIT[0]}
    DBAAS_SERVICE_TYPE=${DBAAS_ENTRY_SPLIT[1]}
    if [ "$DBAAS_SERVICE_NAME" == "$SERVICE_NAME" ]; then
      if [ "$SERVICE_TYPE" == "mariadb" ]; then
        SERVICE_TYPE=$DBAAS_SERVICE_TYPE
      fi
      if [ "$SERVICE_TYPE" == "postgres" ]; then
        SERVICE_TYPE=$DBAAS_SERVICE_TYPE
      fi
      if [ "$SERVICE_TYPE" == "mongo" ]; then
        SERVICE_TYPE=$DBAAS_SERVICE_TYPE
      fi
    fi
  done

  # Do not handle images for shared services
  if  [[ "$SERVICE_TYPE" != "mariadb-dbaas" ]] &&
      [[ "$SERVICE_TYPE" != "mariadb-shared" ]] &&
      [[ "$SERVICE_TYPE" != "postgres-shared" ]] &&
      [[ "$SERVICE_TYPE" != "postgres-dbaas" ]] &&
      [[ "$SERVICE_TYPE" != "mongodb-dbaas" ]] &&
      [[ "$SERVICE_TYPE" != "mongodb-shared" ]]; then
    # Generate list of images to build
    IMAGES+=("${IMAGE_NAME}")
  fi

  # Map Deployment ServiceType to the ImageName
  MAP_DEPLOYMENT_SERVICETYPE_TO_IMAGENAME["${SERVICE_NAME}:${DEPLOYMENT_SERVICETYPE}"]="${IMAGE_NAME}"

  # Create an array with all Service Names and Types if it does not exist yet
  if [[ ! " ${SERVICE_TYPES[@]} " =~ " ${SERVICE_NAME}:${SERVICE_TYPE} " ]]; then
    SERVICE_TYPES+=("${SERVICE_NAME}:${SERVICE_TYPE}")
  fi

  # ServiceName and Type to Original Service Name Mapping, but only once per Service name and Type,
  # as we have original services that appear twice (like in the case of nginx-php)
  if [[ ! "${MAP_SERVICE_TYPE_TO_COMPOSE_SERVICE["${SERVICE_NAME}:${SERVICE_TYPE}"]+isset}" ]]; then
    MAP_SERVICE_TYPE_TO_COMPOSE_SERVICE["${SERVICE_NAME}:${SERVICE_TYPE}"]="${COMPOSE_SERVICE}"
  fi

  # ServiceName to ImageName mapping, but only once as we have original services that appear twice (like in the case of nginx-php)
  # these will be handled via MAP_DEPLOYMENT_SERVICETYPE_TO_IMAGENAME
  if [[ ! "${MAP_SERVICE_NAME_TO_IMAGENAME["${SERVICE_NAME}"]+isset}" ]]; then
    MAP_SERVICE_NAME_TO_IMAGENAME["${SERVICE_NAME}"]="${IMAGE_NAME}"
  fi

done

# Get the pre-rollout and post-rollout vars
LAGOON_PREROLLOUT_DISABLED=$(apiEnvVarCheck LAGOON_PREROLLOUT_DISABLED "false")
LAGOON_POSTROLLOUT_DISABLED=$(apiEnvVarCheck LAGOON_POSTROLLOUT_DISABLED "false")

currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
patchBuildStep "${buildStartTime}" "${buildStartTime}" "${currentStepEnd}" "${NAMESPACE}" "configureVars" "Configure Variables" "false"
previousStepEnd=${currentStepEnd}
beginBuildStep "Container Registry Login" "registryLogin"

##############################################
### CONTAINER REGISTRY LOGIN
##############################################

# $REGISTRY is set as the fallback unauthenticated registry. For use when harbor
# isn't available, like locally or in CI.
REGISTRY="$REGISTRY"

# The internal container registry is configured in lagoon-remote and will be set
# when an authenticated registry, like harbor, is available.
INTERNAL_REGISTRY_URL=$(internalContainerRegistryCheck INTERNAL_REGISTRY_URL)
INTERNAL_REGISTRY_USERNAME=$(internalContainerRegistryCheck INTERNAL_REGISTRY_USERNAME)
INTERNAL_REGISTRY_PASSWORD=$(internalContainerRegistryCheck INTERNAL_REGISTRY_PASSWORD)
if [ -n "$INTERNAL_REGISTRY_URL" ] ; then
  if [ -n "$INTERNAL_REGISTRY_USERNAME" ] && [ -n "$INTERNAL_REGISTRY_PASSWORD" ] ; then
    echo "Logging in to Lagoon main registry"
    docker login -u "$INTERNAL_REGISTRY_USERNAME" -p "$INTERNAL_REGISTRY_PASSWORD" "$INTERNAL_REGISTRY_URL"

    # The $REGISTRY env var is used by the generator, set it to match the internal registry.
    REGISTRY="$INTERNAL_REGISTRY_URL"
  else
    echo "Could not log in to Lagoon main registry"
    if [ -z "$INTERNAL_REGISTRY_USERNAME" ]; then
      echo "No token created for registry ${INTERNAL_REGISTRY_URL}";
    fi
    if [ -z "$INTERNAL_REGISTRY_PASSWORD" ]; then
      echo "No password retrieved for token ${INTERNAL_REGISTRY_USERNAME} in registry ${INTERNAL_REGISTRY_URL}";
    fi

    currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
    patchBuildStep "${buildStartTime}" "${previousStepEnd}" "${currentStepEnd}" "${NAMESPACE}" "registryLogin" "Container Registry Login" "false"
    previousStepEnd=${currentStepEnd}
    exit 1;
  fi
else
  echo "Using unauthenticated registry"
fi

# Generates information needed to build containers:
# - BuildKit enabled/disabled
# - List of "push images"
# - List of "force pull" images
# - Build args
# - Private container registries
ENVIRONMENT_IMAGE_BUILD_DATA=$(build-deploy-tool identify image-builds)

# Private container registries can be configured in Lagoon projects to allow
# pulling private images. If any were set, log in to them now.
for PCR in $(echo "$ENVIRONMENT_IMAGE_BUILD_DATA" | jq -c '.containerRegistries[]? | @base64')
do
  PRIVATE_CONTAINER_REGISTRY=$(echo "${PCR}" | jq -rc '@base64d')
  PRIVATE_CONTAINER_REGISTRY_URL=$(echo "$PRIVATE_CONTAINER_REGISTRY" | jq -r '.url // false')
  PRIVATE_CONTAINER_REGISTRY_CREDENTIAL_USERNAME=$(echo "$PRIVATE_CONTAINER_REGISTRY" | jq -r '.username // false')
  PRIVATE_CONTAINER_REGISTRY_USERNAME_SOURCE=$(echo "$PRIVATE_CONTAINER_REGISTRY" | jq -r '.usernameSource // false')
  PRIVATE_REGISTRY_CREDENTIAL=$(echo "$PRIVATE_CONTAINER_REGISTRY" | jq -r '.password // false')
  PRIVATE_REGISTRY_CREDENTIAL_SOURCE=$(echo "$PRIVATE_CONTAINER_REGISTRY" | jq -r '.passwordSource // false')
  PRIVATE_CONTAINER_REGISTRY_ISDOCKERHUB=$(echo "$PRIVATE_CONTAINER_REGISTRY" | jq -r '.isDockerHub // false')
  if [ $PRIVATE_CONTAINER_REGISTRY_ISDOCKERHUB == "false" ]; then
      echo "Attempting to log in to $PRIVATE_CONTAINER_REGISTRY_URL with user $PRIVATE_CONTAINER_REGISTRY_CREDENTIAL_USERNAME from $PRIVATE_CONTAINER_REGISTRY_USERNAME_SOURCE"
      echo "Using password sourced from $PRIVATE_REGISTRY_CREDENTIAL_SOURCE"
      docker login --username $PRIVATE_CONTAINER_REGISTRY_CREDENTIAL_USERNAME --password $PRIVATE_REGISTRY_CREDENTIAL $PRIVATE_CONTAINER_REGISTRY_URL
  else
      echo "Attempting to log in to docker hub with user $PRIVATE_CONTAINER_REGISTRY_CREDENTIAL_USERNAME from $PRIVATE_CONTAINER_REGISTRY_USERNAME_SOURCE"
      echo "Using password sourced from $PRIVATE_REGISTRY_CREDENTIAL_SOURCE"
      docker login --username $PRIVATE_CONTAINER_REGISTRY_CREDENTIAL_USERNAME --password $PRIVATE_REGISTRY_CREDENTIAL
  fi
done

currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
patchBuildStep "${buildStartTime}" "${buildStartTime}" "${currentStepEnd}" "${NAMESPACE}" "registryLogin" "Container Registry Login" "false"
previousStepEnd=${currentStepEnd}
beginBuildStep "Image Builds" "buildingImages"


##############################################
### BUILD IMAGES
##############################################

for IMAGE_BUILD_DATA in $(echo "$ENVIRONMENT_IMAGE_BUILD_DATA" | jq -c '.images[]')
do
  SERVICE_NAME=$(echo "$IMAGE_BUILD_DATA" | jq -r '.name // false')
  # add the image name to the array of images to push. this is consumed later in the build process
  IMAGES_PUSH["${SERVICE_NAME}"]="$(echo "$IMAGE_BUILD_DATA" | jq -r '.imageBuild.buildImage')"
  if [ "$BUILD_TYPE" == "promote" ]; then
    # add the image name to the array of images to promote from. this is consumed later in the build process
    IMAGES_PROMOTE["${SERVICE_NAME}"]="$(echo "$IMAGE_BUILD_DATA" | jq -r '.imageBuild.promoteImage')"
  fi
done

# we only need to build images for pullrequests and branches
if [[ "$BUILD_TYPE" == "pullrequest"  ||  "$BUILD_TYPE" == "branch" ]]; then
  # use the build-deploy-tool to seed the image build information
  BUILD_ARGS=() # build args are now calculated in the build-deploy tool in the generator step
  # this loop extracts the build arguments from the response from the build deploy tools previous identify image-builds call
  for IMAGE_BUILD_ARGUMENTS in $(echo "$ENVIRONMENT_IMAGE_BUILD_DATA" | jq -r '.buildArguments | to_entries[] | @base64'); do
    BUILD_ARG_NAME=$(echo "$IMAGE_BUILD_ARGUMENTS" | jq -Rr '@base64d | fromjson | .key')
    BUILD_ARG_VALUE=$(echo "$IMAGE_BUILD_ARGUMENTS" | jq -Rr '@base64d | fromjson | .value')
    BUILD_ARGS+=(--build-arg ${BUILD_ARG_NAME}="${BUILD_ARG_VALUE}")
  done

  # Here we iterate over any lagoon.base.image data that has been passed to us
  # in order to explicitly pull the images to ensure they are current
  for FPI in $(echo "$ENVIRONMENT_IMAGE_BUILD_DATA" | jq -rc '.forcePullImages[]?')
  do
    echo "Pulling Image: ${FPI}"
    docker pull "${FPI}"
  done

  # now we loop through the images in the build data and determine if they need to be pulled or build
  for IMAGE_BUILD_DATA in $(echo "$ENVIRONMENT_IMAGE_BUILD_DATA" | jq -c '.images[]')
  do
    SERVICE_NAME=$(echo "$IMAGE_BUILD_DATA" | jq -r '.name // false')
    DOCKERFILE=$(echo "$IMAGE_BUILD_DATA" | jq -r '.imageBuild.dockerFile // false')
    # if there is no dockerfile, then this image needs to be pulled from somewhere else
    if [ $DOCKERFILE == "false" ]; then
      PULL_IMAGE=$(echo "$IMAGE_BUILD_DATA" | jq -r '.imageBuild.pullImage // false')
      if [ "$PULL_IMAGE" != "false" ]; then
        IMAGES_PULL["${SERVICE_NAME}"]="${PULL_IMAGE}"
      fi
    else
      # otherwise extract build information from the image build data payload
      # this is a temporary image name to use for the build, it is based on the namespace and service, this can probably be deprecated and the images could just be
      # built with the name they are meant to be. only 1 build can run at a time within a namespace
      # the temporary name would clash here as well if there were multiple builds (it could use the `imageBuild.buildImage` value)
      TEMPORARY_IMAGE_NAME=$(echo "$IMAGE_BUILD_DATA" | jq -r '.imageBuild.temporaryImage // false')
      # the context for this image build, the original source for this value is from the `docker-compose file`
      BUILD_CONTEXT=$(echo "$IMAGE_BUILD_DATA" | jq -r '.imageBuild.context // ""')
      # the build target for this image build, the original source for this value is from the `docker-compose file`
      BUILD_TARGET=$(echo "$IMAGE_BUILD_DATA" | jq -r '.imageBuild.target // false')
      # determine if buildkit should be disabled for this build
      DOCKER_BUILDKIT=1
      if [ "$(echo "${ENVIRONMENT_IMAGE_BUILD_DATA}" | jq -r '.buildKit')" == "false" ]; then
        DOCKER_BUILDKIT=0
        echo "Not using BuildKit for $DOCKERFILE"
      else
        echo "Using BuildKit for $DOCKERFILE"
      fi
      export DOCKER_BUILDKIT
      BUILD_TARGET_ARGS=""
      if [ $BUILD_TARGET == "false" ]; then
        echo "Building ${BUILD_CONTEXT}/${DOCKERFILE}"
      else
        echo "Building target ${BUILD_TARGET} for ${BUILD_CONTEXT}/${DOCKERFILE}"
        BUILD_TARGET_ARGS="--target ${BUILD_TARGET}"
      fi
      # now do the actual image build, this pipes to tee so that the build output is still realtime in any logs 
      # ie, if someone was looking at the build container logs in k8s
      # this also captures any errors that this command will encounter so that the process can then check the output file to see if the
      # error condition we are looking for is there
      set +e
      (docker build --network=host "${BUILD_ARGS[@]}" -t $TEMPORARY_IMAGE_NAME -f $BUILD_CONTEXT/$DOCKERFILE $BUILD_TARGET_ARGS $BUILD_CONTEXT 2>&1 | tee /kubectl-build-deploy/log-$TEMPORARY_IMAGE_NAME; exit ${PIPESTATUS[0]})
      buildExit=$?
      set -e
      if [ "${buildExit}" != "0" ]; then
        # if the build errors and contains the message we are looking for, then it is probably a buildkit related failure
        # attempt to run run again with --no-cache so that it forces layer invalidation. this will make the build slower, but hopefully succeed
        # why this happens is still to be determined. there isn't enough information in the error to be able to know which layers are the problem
        # or what the actual cause is, making it incredibly difficult to reproduce
        # without being able to reproduce we have to use this workaround to retry :'(
        capErr=0
        if cat /kubectl-build-deploy/log-$TEMPORARY_IMAGE_NAME | grep -q "ERROR: failed to solve: layer does not exist"; then
          capErr=1
        elif cat /kubectl-build-deploy/log-$TEMPORARY_IMAGE_NAME | grep -q "ERROR: failed to solve: failed to prepare"; then
          capErr=1
        elif cat /kubectl-build-deploy/log-$TEMPORARY_IMAGE_NAME | grep -q "ERROR: failed to solve: failed to get layer"; then
          capErr=1
        fi
        if [ "${capErr}" != "0" ]; then
          # at least drop a message saying that this was encountered
          echo "##############################################
The first attempt to build ${BUILD_CONTEXT}/${DOCKERFILE} failed due to a layer error
Retrying build for ${BUILD_CONTEXT}/${DOCKERFILE} without cache
##############################################"
          docker build --no-cache --network=host "${BUILD_ARGS[@]}" -t $TEMPORARY_IMAGE_NAME -f $BUILD_CONTEXT/$DOCKERFILE $BUILD_TARGET_ARGS $BUILD_CONTEXT
        else
          # if the failure is not one that matches the buildkit layer issue, then exit as a normal build failure
          exit 1
        fi
      fi

      # Keep a list of the images we have built, as we need to push them to the registry later
      IMAGES_BUILD["${SERVICE_NAME}"]="${TEMPORARY_IMAGE_NAME}"

      # adding the build image to the list of arguments passed into the next image builds
      SERVICE_NAME_UPPERCASE=$(echo "$SERVICE_NAME" | tr '[:lower:]' '[:upper:]')
    fi
  done
fi

# print information about built image sizes
function printBytes {
    local -i bytes=$1;
    echo "$(( (bytes + 1000000)/1000000 ))MB"
}
if [[ "${IMAGES_BUILD[@]}" ]]; then
  echo "##############################################"
  echo "Built image sizes:"
  echo "##############################################"
fi
for IMAGE_NAME in "${!IMAGES_BUILD[@]}"
do
  TEMPORARY_IMAGE_NAME="${IMAGES_BUILD[${IMAGE_NAME}]}"
  echo -e "Image ${TEMPORARY_IMAGE_NAME}\t\t$(printBytes $(docker inspect ${TEMPORARY_IMAGE_NAME} | jq -r '.[0].Size'))"
done

currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
patchBuildStep "${buildStartTime}" "${previousStepEnd}" "${currentStepEnd}" "${NAMESPACE}" "imageBuildComplete" "Image Builds" "false"
previousStepEnd=${currentStepEnd}
beginBuildStep "Service Configuration Phase" "serviceConfigurationPhase"

##############################################
### CONFIGURE SERVICES, AUTOGENERATED ROUTES AND DBAAS CONFIG
##############################################

YAML_FOLDER="/kubectl-build-deploy/lagoon/services-routes"
mkdir -p $YAML_FOLDER

# BC for routes.insecure, which is now called routes.autogenerate.insecure
BC_ROUTES_AUTOGENERATE_INSECURE=$(cat .lagoon.yml | yq -o json | jq -r '.routes.insecure // false')
if [ ! $BC_ROUTES_AUTOGENERATE_INSECURE == "false" ]; then
  echo "=== routes.insecure is now defined in routes.autogenerate.insecure, pleae update your .lagoon.yml file"
  # update the .lagoon.yml with the new location for build-deploy-tool to read
  yq -i '.routes.autogenerate.insecure = "'${BC_ROUTES_AUTOGENERATE_INSECURE}'"' .lagoon.yml
fi

##############################################
### CREATE SERVICES, AUTOGENERATED ROUTES AND DBAAS CONFIG
##############################################

# generate the autogenerated ingress
AUTOGEN_ROUTES_DISABLED=$(apiEnvVarCheck LAGOON_AUTOGEN_ROUTES_DISABLED false)
if [ ! "$AUTOGEN_ROUTES_DISABLED" == true ]; then
  LAGOON_AUTOGEN_YAML_FOLDER="/kubectl-build-deploy/lagoon/autogen-routes"
  mkdir -p $LAGOON_AUTOGEN_YAML_FOLDER
  build-deploy-tool template autogenerated-ingress --saved-templates-path ${LAGOON_AUTOGEN_YAML_FOLDER}

  # apply autogenerated ingress
  if [ -n "$(ls -A $LAGOON_AUTOGEN_YAML_FOLDER/ 2>/dev/null)" ]; then
    find $LAGOON_AUTOGEN_YAML_FOLDER -type f -exec cat {} \;
    kubectl apply -n ${NAMESPACE} -f $LAGOON_AUTOGEN_YAML_FOLDER/
  fi
else
  echo ">> Autogenerated ingress templates disabled for this build"
# end custom route
fi

# identify any autognerated resources based on their resource name
AUTOGEN_INGRESS=$(build-deploy-tool identify created-ingress | jq -r '.autogenerated[]')
AUTOGEN_ROUTES=$(kubectl -n ${NAMESPACE} get ingress --no-headers -l "lagoon.sh/autogenerated=true" | cut -d " " -f 1 | xargs)
MATCHED_AUTOGEN=false
DELETE_AUTOGEN=()
for AR in $AUTOGEN_ROUTES; do
  for AI in $AUTOGEN_INGRESS; do
    if [ "${AR}" == "${AI}" ]; then
      MATCHED_AUTOGEN=true
      continue
    fi
  done
  if [ "${MATCHED_AUTOGEN}" != "true" ]; then
    DELETE_AUTOGEN+=($AR)
  fi
  MATCHED_AUTOGEN=false
done
for DA in ${!DELETE_AUTOGEN[@]}; do
  # delete any autogenerated ingress in the namespace as they are disabled
  if kubectl -n ${NAMESPACE} get ingress ${DELETE_AUTOGEN[$DA]} &> /dev/null; then
    echo ">> Removing autogenerated ingress for ${DELETE_AUTOGEN[$DA]} because it was disabled"
    kubectl -n ${NAMESPACE} delete ingress ${DELETE_AUTOGEN[$DA]}
  fi
done

for SERVICE_TYPES_ENTRY in "${SERVICE_TYPES[@]}"
do
  echo "=== BEGIN route processing for service ${SERVICE_TYPES_ENTRY} ==="
  IFS=':' read -ra SERVICE_TYPES_ENTRY_SPLIT <<< "$SERVICE_TYPES_ENTRY"

  TEMPLATE_PARAMETERS=()

  SERVICE_NAME=${SERVICE_TYPES_ENTRY_SPLIT[0]}
  SERVICE_TYPE=${SERVICE_TYPES_ENTRY_SPLIT[1]}

  touch /kubectl-build-deploy/${SERVICE_NAME}-values.yaml

done

# generate the dbaas templates if any
LAGOON_DBAAS_YAML_FOLDER="/kubectl-build-deploy/lagoon/dbaas"
mkdir -p $LAGOON_DBAAS_YAML_FOLDER
build-deploy-tool template dbaas --saved-templates-path ${LAGOON_DBAAS_YAML_FOLDER}

# apply dbaas
if [ -n "$(ls -A $LAGOON_DBAAS_YAML_FOLDER/ 2>/dev/null)" ]; then
  find $LAGOON_DBAAS_YAML_FOLDER -type f -exec cat {} \;
  kubectl apply -n ${NAMESPACE} -f $LAGOON_DBAAS_YAML_FOLDER/
fi

currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
patchBuildStep "${buildStartTime}" "${previousStepEnd}" "${currentStepEnd}" "${NAMESPACE}" "serviceConfigurationComplete" "Service Configuration Phase" "false"
previousStepEnd=${currentStepEnd}
beginBuildStep "Route/Ingress Configuration" "configuringRoutes"

TEMPLATE_PARAMETERS=()

##############################################
### CUSTOM ROUTES
##############################################

CUSTOM_ROUTES_DISABLED=$(apiEnvVarCheck LAGOON_CUSTOM_ROUTES_DISABLED false)
if [ ! "$CUSTOM_ROUTES_DISABLED" == true ]; then
  LAGOON_ROUTES_YAML_FOLDER="/kubectl-build-deploy/lagoon/routes"
  mkdir -p $LAGOON_ROUTES_YAML_FOLDER
  build-deploy-tool template ingress --saved-templates-path ${LAGOON_ROUTES_YAML_FOLDER}

  # apply the routes
  if [ -n "$(ls -A $LAGOON_ROUTES_YAML_FOLDER/ 2>/dev/null)" ]; then
    find $LAGOON_ROUTES_YAML_FOLDER -type f -exec cat {} \;
    kubectl apply -n ${NAMESPACE} -f $LAGOON_ROUTES_YAML_FOLDER/
  fi
else
  echo ">> Custom ingress templates disabled for this build"
  # end custom route
fi

currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
patchBuildStep "${buildStartTime}" "${previousStepEnd}" "${currentStepEnd}" "${NAMESPACE}" "configuringRoutesComplete" "Route/Ingress Configuration" "false"
previousStepEnd=${currentStepEnd}
beginBuildStep "Route/Ingress Cleanup" "cleanupRoutes"

##############################################
### CLEANUP Ingress/routes which have been removed from .lagoon.yml
##############################################s

# collect the current routes excluding any certmanager requests.
# its also possible to exclude ingress by adding a label 'route.lagoon.sh/remove=false', this is then used to skip this from the removal checks
CURRENT_ROUTES=$(kubectl -n ${NAMESPACE} get ingress  -l "lagoon.sh/autogenerated!=true"  --no-headers  2> /dev/null | cut -d " " -f 1 | xargs)
# since label selectors can't be combined properly, this is done so that the build can get all the routes
# and then remove any that match our conditions to be ignored by the removal checker
IGNORE_ROUTES=$(kubectl -n ${NAMESPACE} get ingress --no-headers -l "acme.cert-manager.io/http01-solver=true"  2> /dev/null | cut -d " " -f 1 | xargs)
for SINGLE_ROUTE in ${IGNORE_ROUTES}; do
  # remove ignored routes from the current routes
  CURRENT_ROUTES=( "${CURRENT_ROUTES[@]/$SINGLE_ROUTE}" )
done
IGNORE_ROUTES=$(kubectl -n ${NAMESPACE} get ingress --no-headers -l "lagoon.sh/remove=false"  2> /dev/null | cut -d " " -f 1 | xargs)
for SINGLE_ROUTE in ${IGNORE_ROUTES}; do
  # remove ignored routes from the current routes
  CURRENT_ROUTES=( "${CURRENT_ROUTES[@]/$SINGLE_ROUTE}" )
done

# collect the routes that Lagoon thinks it should have based on the .lagoon.yml and any routes that have come from the api
# using the build-deploy-tool generator
YAML_ROUTES_TO_JSON=$(build-deploy-tool identify created-ingress | jq -r '.secondary[]')

MATCHED_INGRESS=false
DELETE_INGRESS=()
# loop over the routes from kubernetes
for SINGLE_ROUTE in ${CURRENT_ROUTES}; do
  # loop over the routes that Lagoon thinks it should have
  for YAML_ROUTE in ${YAML_ROUTES_TO_JSON}; do
    if [ "${SINGLE_ROUTE}" == "${YAML_ROUTE}" ]; then
      MATCHED_INGRESS=true
      continue
    fi
  done
  if [ "${MATCHED_INGRESS}" != "true" ]; then
    DELETE_INGRESS+=($SINGLE_ROUTE)
  fi
  MATCHED_INGRESS=false
done

CLEANUP_WARNINGS="false"
if [ ${#DELETE_INGRESS[@]} -ne 0 ]; then
  CLEANUP_WARNINGS="true"
  ((++BUILD_WARNING_COUNT))
  echo ">> Lagoon detected routes that have been removed from the .lagoon.yml or Lagoon API"
  echo "> If you need these routes, you should update your .lagoon.yml file and make sure the routes exist."
  if [ "$(featureFlag CLEANUP_REMOVED_LAGOON_ROUTES)" != enabled ]; then
    echo "> If you no longer need these routes, you can instruct Lagoon to remove it from the environment by setting the following variable"
    echo "> 'LAGOON_FEATURE_FLAG_CLEANUP_REMOVED_LAGOON_ROUTES=enabled' as a GLOBAL scoped variable to this environment or project"
    echo "> You should remove this variable after the deployment has been completed, otherwise future route removals will happen automatically"
  else
    echo "> 'LAGOON_FEATURE_FLAG_CLEANUP_REMOVED_LAGOON_ROUTES=enabled' is configured and the following routes will be removed."
    echo "> You should remove this variable if you don't want routes to be removed automatically"
  fi
  echo "> Future releases of Lagoon may remove routes automatically, you should ensure that your routes are up always up to date if you see this warning"
  for DI in ${DELETE_INGRESS[@]}
  do
    if [ "$(featureFlag CLEANUP_REMOVED_LAGOON_ROUTES)" = enabled ]; then
      if kubectl -n ${NAMESPACE} get ingress ${DI} &> /dev/null; then
        echo ">> Removing ingress ${DI}"
        kubectl -n ${NAMESPACE} delete ingress ${DI}
        #delete anything else?
      fi
    else
      echo "> The route '${DI}' would be removed"
    fi
  done
else
  echo "No route cleanup required"
fi

currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
patchBuildStep "${buildStartTime}" "${previousStepEnd}" "${currentStepEnd}" "${NAMESPACE}" "routeCleanupComplete" "Route/Ingress Cleanup" "${CLEANUP_WARNINGS}"

##############################################
### Report any ingress that have stale or stalled acme challenges, this accordion will only show if there are stale challenges
##############################################s
# collect any current challenge routes in the namespace that are older than 1 hour (to ignore current build ones or pending ones)
CURRENT_CHALLENGE_ROUTES=$(kubectl -n ${NAMESPACE} get ingress -l "acme.cert-manager.io/http01-solver=true" 2> /dev/null | awk 'match($7,/[0-9]+d|[0-9]+h|[0-9][0-9][0-9]m|[6-9][0-9]m/) {print $1}')
if [ "${CURRENT_CHALLENGE_ROUTES[@]}" != "" ]; then
  previousStepEnd=${currentStepEnd}
  beginBuildStep "Route/Ingress Certificate Challenges" "staleChallenges"
  ((++BUILD_WARNING_COUNT))
  echo ">> Lagoon detected routes that have stale acme certificate challenges."
  echo "  This indicates that the routes have not generated the certificate for some reason."
  echo "  You may need to verify that the DNS or configuration is correct for the hosting provider."
  echo "  ${LAGOON_FEATURE_FLAG_DEFAULT_DOCUMENTATION_URL}/using-lagoon-the-basics/going-live/#routes-ssl"
  echo "  Depending on your going live instructions from your hosting provider, you may need to make adjustments to your .lagoon.yml file"
  echo "  Otherwise, If you no longer need these routes, you should remove them from your .lagoon.yml file."
  echo ""
  for CR in ${CURRENT_CHALLENGE_ROUTES[@]}
  do
      echo ">> The route '${CR}' has stale certificate challenge"
      # grab the error after 'order is' because the pretext could lead to confusion
      FAILURE_REASON=$(kubectl -n ${NAMESPACE} get certificate.cert-manager.io ${CR}-tls -o json | jq -r '.status.conditions[] | select (.reason=="Failed") | .message' | grep -oP "order is.*$")
      if [ -z "$FAILURE_REASON" ]; then # if there is a capturable failure reason, print it here
        echo "  reason: ${FAILURE_REASON}"
      fi
  done

  currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
  patchBuildStep "${buildStartTime}" "${previousStepEnd}" "${currentStepEnd}" "${NAMESPACE}" "staleChallengesComplete" "Route/Ingress Certificate Challenges" "true"
fi
previousStepEnd=${currentStepEnd}
beginBuildStep "Update Environment Secrets" "updateEnvSecrets"

##############################################
### PROJECT WIDE ENV VARIABLES
##############################################

# identify primary-ingress scans the builds autogenerated and custom ingresses looking for the `main` route
# scans autogen, custom defined, and finally activestandby. first in the list is always returned for each state with each
# step overwriting the previous so only 1 ingress is returned
# previous check looked for `spec.tls` which always exists in our kubernetes templates
# so just add https...
ROUTE=$(build-deploy-tool identify primary-ingress)
if [ ! -z "${ROUTE}" ]; then
  ROUTE=${ROUTE}
fi
# if both route generations are disabled, don't set a route
if [[ "$CUSTOM_ROUTES_DISABLED" == true ]] && [[ "$AUTOGEN_ROUTES_DISABLED" == true ]]; then
  ROUTE=""
fi

# Load all routes with correct schema and comma separated
ROUTES=$(kubectl -n ${NAMESPACE} get ingress --sort-by='{.metadata.name}' -l "acme.cert-manager.io/http01-solver!=true" -o=go-template --template='{{range $indexItems, $ingress := .items}}{{if $indexItems}},{{end}}{{$tls := .spec.tls}}{{range $indexRule, $rule := .spec.rules}}{{if $indexRule}},{{end}}{{if $tls}}https://{{else}}http://{{end}}{{.host}}{{end}}{{end}}')

# swap dioscuri for activestanby label
for ingress in $(kubectl  -n ${NAMESPACE} get ingress -l "dioscuri.amazee.io/migrate" -o json | jq -r '.items[] | @base64'); do
    INGRESS_NAME=$(echo $ingress | jq -Rr '@base64d | fromjson | .metadata.name')
    MIGRATE_VALUE=$(echo $ingress | jq -Rr '@base64d | fromjson | .metadata.labels["dioscuri.amazee.io/migrate"] // false')
    PATCH='{
  "metadata": {
    "labels": {
      "activestandby.lagoon.sh/migrate": "'${MIGRATE_VALUE}'",
      "dioscuri.amazee.io/migrate": null,
      "dioscuri.amazee.io/migrated-from": null
    }
  }
}'
    kubectl -n ${NAMESPACE} patch ingress ${INGRESS_NAME} -p "${PATCH}"
done

# Active / Standby routes
ACTIVE_ROUTES=""
STANDBY_ROUTES=""
if [ ! -z "${STANDBY_ENVIRONMENT}" ]; then
ACTIVE_ROUTES=$(kubectl -n ${NAMESPACE} get ingress --sort-by='{.metadata.name}' -l "activestandby.lagoon.sh/migrate=true" -o=go-template --template='{{range $indexItems, $ingress := .items}}{{if $indexItems}},{{end}}{{$tls := .spec.tls}}{{range $indexRule, $rule := .spec.rules}}{{if $indexRule}},{{end}}{{if $tls}}https://{{else}}http://{{end}}{{.host}}{{end}}{{end}}')
STANDBY_ROUTES=$(kubectl -n ${NAMESPACE} get ingress --sort-by='{.metadata.name}' -l "activestandby.lagoon.sh/migrate=true" -o=go-template --template='{{range $indexItems, $ingress := .items}}{{if $indexItems}},{{end}}{{$tls := .spec.tls}}{{range $indexRule, $rule := .spec.rules}}{{if $indexRule}},{{end}}{{if $tls}}https://{{else}}http://{{end}}{{.host}}{{end}}{{end}}')
fi

# Get list of autogenerated routes
AUTOGENERATED_ROUTES=$(kubectl -n ${NAMESPACE} get ingress --sort-by='{.metadata.name}' -l "lagoon.sh/autogenerated=true" -o=go-template --template='{{range $indexItems, $ingress := .items}}{{if $indexItems}},{{end}}{{$tls := .spec.tls}}{{range $indexRule, $rule := .spec.rules}}{{if $indexRule}},{{end}}{{if $tls}}https://{{else}}http://{{end}}{{.host}}{{end}}{{end}}')

# loop through created DBAAS templates
DBAAS=($(build-deploy-tool identify dbaas))
for DBAAS_ENTRY in "${DBAAS[@]}"
do
  IFS=':' read -ra DBAAS_ENTRY_SPLIT <<< "$DBAAS_ENTRY"

  SERVICE_NAME=${DBAAS_ENTRY_SPLIT[0]}
  SERVICE_TYPE=${DBAAS_ENTRY_SPLIT[1]}
  SERVICE_NAME_UPPERCASE=$(echo "$SERVICE_NAME" | tr '[:lower:]' '[:upper:]' | tr '-' '_')
  if [[ "$SERVICE_TYPE}" =~ "-single" ]]; then
    # skip to next if this type is a single
    continue
  fi
  case "$SERVICE_TYPE" in

    mariadb-dbaas)
        # remove the image from images to pull
        unset IMAGES_PULL[$SERVICE_NAME]
        CONSUMER_TYPE="mariadbconsumer"
        . /kubectl-build-deploy/scripts/exec-kubectl-dbaas-wait.sh
        MARIADB_DBAAS_CONSUMER_SPECS["${SERVICE_NAME}"]=$(kubectl -n ${NAMESPACE} get mariadbconsumer/${SERVICE_NAME} -o json | jq -r '.spec | @base64')
        ;;

    postgres-dbaas)
        # remove the image from images to pull
        unset IMAGES_PULL[$SERVICE_NAME]
        CONSUMER_TYPE="postgresqlconsumer"
        . /kubectl-build-deploy/scripts/exec-kubectl-dbaas-wait.sh
        POSTGRES_DBAAS_CONSUMER_SPECS["${SERVICE_NAME}"]=$(kubectl -n ${NAMESPACE} get postgresqlconsumer/${SERVICE_NAME} -o json | jq -r '.spec | @base64')
        ;;

    mongodb-dbaas)
        # remove the image from images to pull
        unset IMAGES_PULL[$SERVICE_NAME]
        CONSUMER_TYPE="mongodbconsumer"
        . /kubectl-build-deploy/scripts/exec-kubectl-dbaas-wait.sh
        MONGODB_DBAAS_CONSUMER_SPECS["${SERVICE_NAME}"]=$(kubectl -n ${NAMESPACE} get mongodbconsumer/${SERVICE_NAME} -o json | jq -r '.spec | @base64')
        ;;

    *)
        echo "DBAAS Type ${SERVICE_TYPE} not implemented"; exit 1;

  esac
done

# convert specs into credential dump for ingestion by build-deploy-tool
DBAAS_VARIABLES="[]"
for SERVICE_NAME in "${!MARIADB_DBAAS_CONSUMER_SPECS[@]}"
do
  SERVICE_NAME_UPPERCASE=$(echo "$SERVICE_NAME" | tr '[:lower:]' '[:upper:]' | tr '-' '_')
  DB_HOST=$(echo ${MARIADB_DBAAS_CONSUMER_SPECS["$SERVICE_NAME"]} | jq -Rr '@base64d | fromjson | .consumer.services.primary')
  DB_USER=$(echo ${MARIADB_DBAAS_CONSUMER_SPECS["$SERVICE_NAME"]} | jq -Rr '@base64d | fromjson | .consumer.username')
  DB_PASSWORD=$(echo ${MARIADB_DBAAS_CONSUMER_SPECS["$SERVICE_NAME"]} | jq -Rr '@base64d | fromjson | .consumer.password')
  DB_NAME=$(echo ${MARIADB_DBAAS_CONSUMER_SPECS["$SERVICE_NAME"]} | jq -Rr '@base64d | fromjson | .consumer.database')
  DB_PORT=$(echo ${MARIADB_DBAAS_CONSUMER_SPECS["$SERVICE_NAME"]} | jq -Rr '@base64d | fromjson | .provider.port')
  DB_CONSUMER='{"'${SERVICE_NAME_UPPERCASE}'_HOST":"'${DB_HOST}'", "'${SERVICE_NAME_UPPERCASE}'_USERNAME":"'${DB_USER}'","'${SERVICE_NAME_UPPERCASE}'_PASSWORD":"'${DB_PASSWORD}'","'${SERVICE_NAME_UPPERCASE}'_DATABASE":"'${DB_NAME}'","'${SERVICE_NAME_UPPERCASE}'_PORT":"'${DB_PORT}'"}'
  if DB_READREPLICA_HOSTS=$(echo ${MARIADB_DBAAS_CONSUMER_SPECS["$SERVICE_NAME"]} | jq -Rr '@base64d | fromjson | .consumer.services.replicas | .[]' 2>/dev/null); then
    if [ "$DB_READREPLICA_HOSTS" != "null" ]; then
      DB_READREPLICA_HOSTS=$(echo "$DB_READREPLICA_HOSTS" | sed 's/^\|$//g' | paste -sd, -)
      DB_CONSUMER=$(echo "${DB_CONSUMER}" | jq '. + {"'${SERVICE_NAME_UPPERCASE}'_READREPLICA_HOSTS":"'${DB_READREPLICA_HOSTS}'"}')
    fi
  fi
  DBAAS_VARIABLES=$(echo "$DBAAS_VARIABLES" | jq '. + '$(echo "$DB_CONSUMER" | jq -sMrc)'')
done

for SERVICE_NAME in "${!POSTGRES_DBAAS_CONSUMER_SPECS[@]}"
do
  SERVICE_NAME_UPPERCASE=$(echo "$SERVICE_NAME" | tr '[:lower:]' '[:upper:]' | tr '-' '_')
  DB_HOST=$(echo ${POSTGRES_DBAAS_CONSUMER_SPECS["$SERVICE_NAME"]} | jq -Rr '@base64d | fromjson | .consumer.services.primary')
  DB_USER=$(echo ${POSTGRES_DBAAS_CONSUMER_SPECS["$SERVICE_NAME"]} | jq -Rr '@base64d | fromjson | .consumer.username')
  DB_PASSWORD=$(echo ${POSTGRES_DBAAS_CONSUMER_SPECS["$SERVICE_NAME"]} | jq -Rr '@base64d | fromjson | .consumer.password')
  DB_NAME=$(echo ${POSTGRES_DBAAS_CONSUMER_SPECS["$SERVICE_NAME"]} | jq -Rr '@base64d | fromjson | .consumer.database')
  DB_PORT=$(echo ${POSTGRES_DBAAS_CONSUMER_SPECS["$SERVICE_NAME"]} | jq -Rr '@base64d | fromjson | .provider.port')
  DB_CONSUMER='{"'${SERVICE_NAME_UPPERCASE}'_HOST":"'${DB_HOST}'", "'${SERVICE_NAME_UPPERCASE}'_USERNAME":"'${DB_USER}'","'${SERVICE_NAME_UPPERCASE}'_PASSWORD":"'${DB_PASSWORD}'","'${SERVICE_NAME_UPPERCASE}'_DATABASE":"'${DB_NAME}'","'${SERVICE_NAME_UPPERCASE}'_PORT":"'${DB_PORT}'"}'
  if DB_READREPLICA_HOSTS=$(echo ${POSTGRES_DBAAS_CONSUMER_SPECS["$SERVICE_NAME"]} | jq -Rr '@base64d | fromjson | .consumer.services.replicas | .[]' 2>/dev/null); then
    if [ "$DB_READREPLICA_HOSTS" != "null" ]; then
      DB_READREPLICA_HOSTS=$(echo "$DB_READREPLICA_HOSTS" | sed 's/^\|$//g' | paste -sd, -)
      DB_CONSUMER=$(echo "${DB_CONSUMER}" | jq '. + {"'${SERVICE_NAME_UPPERCASE}'_READREPLICA_HOSTS":"'${DB_READREPLICA_HOSTS}'"}')
    fi
  fi
  DBAAS_VARIABLES=$(echo "$DBAAS_VARIABLES" | jq '. + '$(echo "$DB_CONSUMER" | jq -sMrc)'')
done

for SERVICE_NAME in "${!MONGODB_DBAAS_CONSUMER_SPECS[@]}"
do
  SERVICE_NAME_UPPERCASE=$(echo "$SERVICE_NAME" | tr '[:lower:]' '[:upper:]' | tr '-' '_')
  DB_HOST=$(echo ${MONGODB_DBAAS_CONSUMER_SPECS["$SERVICE_NAME"]} | jq -Rr '@base64d | fromjson | .consumer.services.primary')
  DB_USER=$(echo ${MONGODB_DBAAS_CONSUMER_SPECS["$SERVICE_NAME"]} | jq -Rr '@base64d | fromjson | .consumer.username')
  DB_PASSWORD=$(echo ${MONGODB_DBAAS_CONSUMER_SPECS["$SERVICE_NAME"]} | jq -Rr '@base64d | fromjson | .consumer.password')
  DB_NAME=$(echo ${MONGODB_DBAAS_CONSUMER_SPECS["$SERVICE_NAME"]} | jq -Rr '@base64d | fromjson | .consumer.database')
  DB_PORT=$(echo ${MONGODB_DBAAS_CONSUMER_SPECS["$SERVICE_NAME"]} | jq -Rr '@base64d | fromjson | .provider.port')
  DB_AUTHSOURCE=$(echo ${MONGODB_DBAAS_CONSUMER_SPECS["$SERVICE_NAME"]} | jq -Rr '@base64d | fromjson | .provider.auth.source')
  DB_AUTHMECHANISM=$(echo ${MONGODB_DBAAS_CONSUMER_SPECS["$SERVICE_NAME"]} | jq -Rr '@base64d | fromjson | .provider.auth.mechanism')
  DB_AUTHTLS=$(echo ${MONGODB_DBAAS_CONSUMER_SPECS["$SERVICE_NAME"]} | jq -Rr '@base64d | fromjson | .provider.auth.tls')
  DB_CONSUMER='{"'${SERVICE_NAME_UPPERCASE}'_HOST":"'${DB_HOST}'", "'${SERVICE_NAME_UPPERCASE}'_USERNAME":"'${DB_USER}'", "'${SERVICE_NAME_UPPERCASE}'_PASSWORD":"'${DB_PASSWORD}'", "'${SERVICE_NAME_UPPERCASE}'_DATABASE":"'${DB_NAME}'", "'${SERVICE_NAME_UPPERCASE}'_PORT":"'${DB_PORT}'", "'${SERVICE_NAME_UPPERCASE}'_AUTHSOURCE":"'${DB_AUTHSOURCE}'", "'${SERVICE_NAME_UPPERCASE}'_AUTHMECHANISM":"'${DB_AUTHMECHANISM}'", "'${SERVICE_NAME_UPPERCASE}'_AUTHTLS":"'${DB_AUTHTLS}'"}'
  DBAAS_VARIABLES=$(echo "$DBAAS_VARIABLES" | jq '. + '$(echo "$DB_CONSUMER" | jq -sMrc)'')
done
echo "$DBAAS_VARIABLES" | jq -Mr > /kubectl-build-deploy/dbaas-creds.json

# Generate the lagoon-env secret
LAGOON_ENV_YAML_FOLDER="/kubectl-build-deploy/lagoon/lagoon-env"
mkdir -p $LAGOON_ENV_YAML_FOLDER
# for now, pass the `--routes` flag to the template command so that the routes from the cluster are used in the `lagoon-env` secret LAGOON_ROUTES as this is how it used to be
# since this tool currently has no kube scrape, and the ones the tool generates are only the ones it knows about currently
# we have to source them this way for now. In the future though, we'll be able to omit this flag and remove it from the tool
# also would be part of https://github.com/uselagoon/build-deploy-tool/blob/f527a89ad5efb46e19a2f59d9ff3ffbff541e2a2/legacy/build-deploy-docker-compose.sh#L1090
echo "Updating lagoon-env secret"
build-deploy-tool template lagoon-env \
  --secret-name "lagoon-env" \
  --saved-templates-path ${LAGOON_ENV_YAML_FOLDER} \
  --dbaas-creds /kubectl-build-deploy/dbaas-creds.json \
  --routes "${ROUTES}"
kubectl apply -n ${NAMESPACE} -f ${LAGOON_ENV_YAML_FOLDER}/lagoon-env-secret.yaml

if kubectl -n ${NAMESPACE} get configmap lagoon-env &> /dev/null; then
  # this section will only run once on the initial change from configmap to secret
  # convert the existing configmap into a secret and then remove anything that the API has provided to the `lagoon-env` secret
  # this is going to make it so that anything that isn't in the API is added to a new secret called `lagoon-platform-env` which is where non-api variables can be added
  # by platform operators without impacting the main lagoon-env secret, this is to fix https://github.com/uselagoon/build-deploy-tool/issues/136
  # this will also make it so that if a user has deleted a variable from the api in the past, it will still exist in the lagoon-platform-env secret so that there
  # is no change in behaviour for the user and not seeing unexpectedly a variable they may have deleted they were still relying on
  # unfortunately, variables that remain in the lagoon-platform-env secret will never be deleted
  # this secret may end up being empty if everything in the API is correct and there are no discrepancies.
  CURRENT_CONFIGMAP_VARS=$(kubectl -n ${NAMESPACE} get configmap lagoon-env -o json | jq -cr '.data')
  build-deploy-tool template lagoon-env \
    --secret-name "lagoon-platform-env" \
    --saved-templates-path ${LAGOON_ENV_YAML_FOLDER} \
    --dbaas-creds /kubectl-build-deploy/dbaas-creds.json \
    --configmap-vars "${CURRENT_CONFIGMAP_VARS}" \
    --routes "${ROUTES}"
  kubectl apply -n ${NAMESPACE} -f ${LAGOON_ENV_YAML_FOLDER}/lagoon-platform-env-secret.yaml
  # the old lagoon-env configmap will be removed at the end of the applying deployments step so that in the event of a failure between this point
  # and the rollouts completing, the configmap will still exist if the failure occurs before the deployments are applied
fi
# if the lagoon-platform-env secret doesn't exist, create an empty one
if ! kubectl -n ${NAMESPACE} get secret lagoon-platform-env &> /dev/null; then
  build-deploy-tool template lagoon-env \
    --secret-name "lagoon-platform-env" \
    --saved-templates-path ${LAGOON_ENV_YAML_FOLDER} \
    --dbaas-creds /kubectl-build-deploy/dbaas-creds.json \
    --routes "${ROUTES}"
  kubectl apply -n ${NAMESPACE} -f ${LAGOON_ENV_YAML_FOLDER}/lagoon-platform-env-secret.yaml
fi

# now remove any vars from the lagoon-env secret that were deleted from the API
EXISTING_LAGOONENV_VARS=$(kubectl -n ${NAMESPACE} get secret lagoon-env -o json  2> /dev/null | jq -r '.data | keys[]')
# if there were existing vars in the secret
# work out which ones no longer exist in the API and run patch op remove on them
if [ ! -z "$EXISTING_LAGOONENV_VARS" ]; then
  # get what is in the secret now that the patch operations to add what is in the API has been done already
  CURRENT_LAGOONENV_VARS=$(kubectl -n ${NAMESPACE} get secret lagoon-env -o json | jq -r '.data | keys[]')

  # since the secret we generated at the start contains only variables that are generated by the generator
  # and provided by the lagoon-api, we can use it to work out what to remove from the existing secret
  # since the existing secret could contain variables that aren't in the api, we compare these 2 things to see what needs to be removed from the secret
  CREATED_LAGOONENV_VARS=$(cat ${LAGOON_ENV_YAML_FOLDER}/lagoon-env-secret.yaml | yq -o json | jq -r '.stringData | keys[]')
  VARS_TO_REMOVE=$(comm -23 <(echo $CURRENT_LAGOONENV_VARS | tr ' ' '\n' | sort) <(echo $CREATED_LAGOONENV_VARS | tr ' ' '\n' | sort))

  # now work out the patch operations to remove the unneeded keys from the secret
  REMOVE_OPERATION_JSON=""
  # if there are vars to remove, then craft the remove operation patch
  if [ ! -z "$VARS_TO_REMOVE" ]; then
    for VAR_TO_REMOVE in $VARS_TO_REMOVE
    do
      REMOVE_OPERATION_JSON="${REMOVE_OPERATION_JSON:+$REMOVE_OPERATION_JSON, }$(echo -n {\"op\": \"remove\", \"path\": \"/data/$VAR_TO_REMOVE\"})"
    done
    # then actually apply the patch to remove the vars from the secret
    kubectl patch \
      -n ${NAMESPACE} \
      secret lagoon-env \
      --type=json -p "[$REMOVE_OPERATION_JSON]"
  fi
fi

# do a comparison between what is in the current lagoon-env secret and the lagoon-platform-env secret
# collect the current vars from both secrets
CURRENT_LAGOONPLATFORMENV_VARS=$(kubectl -n ${NAMESPACE} get secret lagoon-platform-env -o json  2> /dev/null | jq -r 'select(.data != null) | .data | keys[]')
CURRENT_LAGOONENV_VARS=$(kubectl -n ${NAMESPACE} get secret lagoon-env -o json  2> /dev/null | jq -r 'select(.data != null) | .data | keys[]')
if [[ ! -z "${CURRENT_LAGOONPLATFORMENV_VARS}" ]] && [[ ! -z "${CURRENT_LAGOONENV_VARS}" ]]; then
  # since the lagoon-platform-env secret is never populated by machine, only human
  # we can check if a user has added a variable that may have previously existed and was deleted from the API has been added again
  # then we can remove it from the `lagoon-platform-env` secret, allowing for the user to delete it again from the API
  # the variable will then correctly get deleted from the `lagoon-env` secret like it should in the step prior to this

  # get variable names present in BOTH secrets, if it exists in both, we need to remove it from the `lagoon-platform-env` secret
  # this will then allow its deletion from the main `lagoon-env` secret if it ever gets deleted from the lagoon api
  # the preference is for variables in the API to exist, rather than being set manually in kubernetes, hence the `lagoon-platform-env` secret remains
  # mostly untouched except to remove variables from if they're ever detected from the lagoon api
  # yes, this means that the value of the variables could be different, but the assumption will be that a user adding the variable to the api
  # assumes they understand what it does, as it would have overwritten a variable in the lagoon-env configmap in the past anyway
  # so this process is just to correct the bug with removing variables from the api should remove them from the secret too
  VARS_TO_REMOVE=$(comm -12 <(echo $CURRENT_LAGOONPLATFORMENV_VARS | tr ' ' '\n' | sort) <(echo $CURRENT_LAGOONENV_VARS | tr ' ' '\n' | sort))
  # now work out the patch operations to remove the unneeded keys from the secret
  REMOVE_OPERATION_JSON=""
  # if there are vars to remove, then craft the remove operation patch
  if [ ! -z "$VARS_TO_REMOVE" ]; then
    for VAR_TO_REMOVE in $VARS_TO_REMOVE
    do
      REMOVE_OPERATION_JSON="${REMOVE_OPERATION_JSON:+$REMOVE_OPERATION_JSON, }$(echo -n {\"op\": \"remove\", \"path\": \"/data/$VAR_TO_REMOVE\"})"
    done
    # then actually apply the patch to remove the vars from the secret
    kubectl patch \
      -n ${NAMESPACE} \
      secret lagoon-platform-env \
      --type=json -p "[$REMOVE_OPERATION_JSON]"
  fi
fi

# display a warning if there are variables present in the `lagoon-platform-env` secret that don't exist in the api
# and instruct the user to either add the variable to the API, or contact support if they are unsure what the variable is
# insert warning message generator here?

currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
patchBuildStep "${buildStartTime}" "${previousStepEnd}" "${currentStepEnd}" "${NAMESPACE}" "updateEnvSecretsComplete" "Update Environment Secrets" "false"
previousStepEnd=${currentStepEnd}
beginBuildStep "Image Push to Registry" "pushingImages"

##############################################
### REDEPLOY DEPLOYMENTS IF CONFIG MAP CHANGES
##############################################

# calculate the combined lagoon-env and lagoon-platform-env sha to determine if changes to any secrets have been made
# which will force the deployments to restart as required
LAGOONENV_SHA=$(kubectl --insecure-skip-tls-verify -n ${NAMESPACE} get secret lagoon-env -o yaml | yq -M '.data' | sha256sum | awk '{print $1}')
LAGOONPLATFORMENV_SHA=$(kubectl --insecure-skip-tls-verify -n ${NAMESPACE} get secret lagoon-platform-env -o yaml | yq -M '.data' | sha256sum | awk '{print $1}')
CONFIG_MAP_SHA=$(echo $LAGOONENV_SHA$LAGOONPLATFORMENV_SHA | sha256sum | awk '{print $1}')
export CONFIG_MAP_SHA

##############################################
### PUSH IMAGES TO REGISTRY
##############################################

# set up image deprecation warnings
DEPRECATED_IMAGE_WARNINGS="false"
declare -A DEPRECATED_IMAGE_NAME
declare -A DEPRECATED_IMAGE_STATUS
declare -A DEPRECATED_IMAGE_SUGGESTION

# pullrequest/branch start
if [ "$BUILD_TYPE" == "pullrequest" ] || [ "$BUILD_TYPE" == "branch" ]; then

  # All images that should be pulled are copied to the harbor registry
  for IMAGE_NAME in "${!IMAGES_PULL[@]}"
  do
    PULL_IMAGE="${IMAGES_PULL[${IMAGE_NAME}]}" #extract the pull image name from the images to pull list
    PUSH_IMAGE="${IMAGES_PUSH[${IMAGE_NAME}]}" #extract the push image name from the images to push list

    # Try to handle private registries first

    # the external pull image name is all calculated in the build-deploy tool now, it knows how to calculate it
    # from being a promote image, or an image from an imagecache or from some other registry entirely
    skopeo copy --retry-times 5 --dest-tls-verify=false docker://${PULL_IMAGE} docker://${PUSH_IMAGE}

    # store the resulting image hash
    SKOPEO_INSPECT=$(skopeo inspect --retry-times 5 docker://${PUSH_IMAGE} --tls-verify=false)
    IMAGE_HASHES[${IMAGE_NAME}]=$(echo "${SKOPEO_INSPECT}" | jq ".Name + \"@\" + .Digest" -r)

    # check if the pull through image is deprecated
    DEPRECATED_STATUS=$(echo "${SKOPEO_INSPECT}" | jq -r '.Labels."sh.lagoon.image.deprecated.status" // false')
    if [ "${DEPRECATED_STATUS}" != false ]; then
      DEPRECATED_IMAGE_WARNINGS="true"
      DEPRECATED_IMAGE_NAME[${IMAGE_NAME}]=${PULL_IMAGE#$IMAGECACHE_REGISTRY}
      DEPRECATED_IMAGE_STATUS[${IMAGE_NAME}]=$DEPRECATED_STATUS
      DEPRECATED_IMAGE_SUGGESTION[${IMAGE_NAME}]=$(echo "${SKOPEO_INSPECT}" | jq -r '.Labels."sh.lagoon.image.deprecated.suggested" | sub("docker.io\/";"")? // false')
    fi
  done

  for IMAGE_NAME in "${!IMAGES_BUILD[@]}"
  do
    PUSH_IMAGE="${IMAGES_PUSH[${IMAGE_NAME}]}" #extract the push image name from the images to push list
    # Before the push the temporary name is resolved to the future tag with the registry in the image name
    TEMPORARY_IMAGE_NAME="${IMAGES_BUILD[${IMAGE_NAME}]}"

    # This will actually not push any images and instead just add them to the file /kubectl-build-deploy/lagoon/push
    # this file is used to perform parallel image pushes next
    docker tag ${TEMPORARY_IMAGE_NAME} ${PUSH_IMAGE}
    echo "docker push ${PUSH_IMAGE}" >> /kubectl-build-deploy/lagoon/push

    # check if the built image is deprecated
    DOCKER_IMAGE_OUTPUT=$(docker inspect ${TEMPORARY_IMAGE_NAME})
    DEPRECATED_STATUS=$(echo "${DOCKER_IMAGE_OUTPUT}" | jq -r '.[] | .Config.Labels."sh.lagoon.image.deprecated.status" // false')
    if [ "${DEPRECATED_STATUS}" != false ]; then
      DEPRECATED_IMAGE_WARNINGS="true"
      DEPRECATED_IMAGE_NAME[${IMAGE_NAME}]=$TEMPORARY_IMAGE_NAME
      DEPRECATED_IMAGE_STATUS[${IMAGE_NAME}]=$DEPRECATED_STATUS
      DEPRECATED_IMAGE_SUGGESTION[${IMAGE_NAME}]=$(echo "${DOCKER_IMAGE_OUTPUT}" | jq -r '.[] | .Config.Labels."sh.lagoon.image.deprecated.suggested" | sub("docker.io\/";"")? // false')
    fi
  done

  # If we have images to push to the registry, let's do so
  if [ -f /kubectl-build-deploy/lagoon/push ]; then
    parallel --retries 4 < /kubectl-build-deploy/lagoon/push
  fi

  # load the image hashes for just pushed images
  for IMAGE_NAME in "${!IMAGES_BUILD[@]}"
  do
    PUSH_IMAGE="${IMAGES_PUSH[${IMAGE_NAME}]}" #extract the push image name from the images to push list
    JQ_QUERY=(jq -r ".[]|select(test(\"${REGISTRY}/${PROJECT}/${ENVIRONMENT}/${IMAGE_NAME}@\"))")
    IMAGE_HASHES[${IMAGE_NAME}]=$(docker inspect ${PUSH_IMAGE} --format '{{json .RepoDigests}}' | "${JQ_QUERY[@]}")
  done

# pullrequest/branch end
# promote start
elif [ "$BUILD_TYPE" == "promote" ]; then

  for IMAGE_NAME in "${IMAGES[@]}"
  do
    PUSH_IMAGE="${IMAGES_PUSH[${IMAGE_NAME}]}" #extract the push image name from the images to push list
    PROMOTE_IMAGE="${IMAGES_PROMOTE[${IMAGE_NAME}]}" #extract the push image name from the images to push list
    skopeo copy --retry-times 5 --src-tls-verify=false --dest-tls-verify=false docker://${PROMOTE_IMAGE} docker://${PUSH_IMAGE}

    IMAGE_HASHES[${IMAGE_NAME}]=$(skopeo inspect --retry-times 5 docker://${PUSH_IMAGE} --tls-verify=false | jq ".Name + \"@\" + .Digest" -r)
  done
# promote end
fi

currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
patchBuildStep "${buildStartTime}" "${previousStepEnd}" "${currentStepEnd}" "${NAMESPACE}" "imagePushComplete" "Image Push to Registry" "false"

##############################################
### Check for deprecated images
##############################################

if [ "${DEPRECATED_IMAGE_WARNINGS}" == "true" ]; then
  previousStepEnd=${currentStepEnd}
  beginBuildStep "Deprecated Image Warnings" "deprecatedImages"
  ((++BUILD_WARNING_COUNT))
  echo ">> Lagoon detected deprecated images during the build"
  echo "  This indicates that an image you're using in the build has been flagged as deprecated."
  echo "  You should stop using these images as soon as possible."
  echo "  If the deprecated image has a suggested replacement, it will be mentioned in the warning."
  echo "  Please visit ${LAGOON_FEATURE_FLAG_DEFAULT_DOCUMENTATION_URL}/deprecated-images for more information."
  echo ""
  for IMAGE_NAME in "${!DEPRECATED_IMAGE_NAME[@]}"
  do
    echo ">> The image (or an image used in the build for) ${DEPRECATED_IMAGE_NAME[${IMAGE_NAME}]} has been deprecated, marked ${DEPRECATED_IMAGE_STATUS[${IMAGE_NAME}]}"
    if [ "${DEPRECATED_IMAGE_SUGGESTION[${IMAGE_NAME}]}" != "false" ]; then
      echo "  A suggested replacement image is ${DEPRECATED_IMAGE_SUGGESTION[${IMAGE_NAME}]}"
    else
      echo "  No replacement image has been suggested"
    fi
    echo ""
  done

  currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
  patchBuildStep "${buildStartTime}" "${previousStepEnd}" "${currentStepEnd}" "${NAMESPACE}" "deprecatedImagesComplete" "Deprecated Image Warnings" "true"
fi

previousStepEnd=${currentStepEnd}
beginBuildStep "Backup Configuration" "configuringBackups"

# Run the backup generation script

BACKUPS_DISABLED=$(apiEnvVarCheck LAGOON_BACKUPS_DISABLED false)
if [ ! "$BACKUPS_DISABLED" == true ]; then
  # check if k8up v2 feature flag is enabled
  LAGOON_BACKUP_YAML_FOLDER="/kubectl-build-deploy/lagoon/backup"
  mkdir -p $LAGOON_BACKUP_YAML_FOLDER
  if [ "$(featureFlag K8UP_V2)" = enabled ]; then
  # build-tool doesn't do any capability checks yet, so do this for now
    if kubectl -n ${NAMESPACE} get schedule.k8up.io &> /dev/null; then
    echo "Backups: generating k8up.io/v1 resources"
      if ! kubectl --insecure-skip-tls-verify -n ${NAMESPACE} get secret baas-repo-pw &> /dev/null; then
        # Create baas-repo-pw secret based on the project secret
        kubectl --insecure-skip-tls-verify -n ${NAMESPACE} create secret generic baas-repo-pw --from-literal=repo-pw=$(echo -n "${PROJECT_SECRET}-BAAS-REPO-PW" | sha256sum | cut -d " " -f 1)
      fi
      build-deploy-tool template backup-schedule --version v2 --saved-templates-path ${LAGOON_BACKUP_YAML_FOLDER}
      # check if the existing schedule exists, and delete it
      if kubectl -n ${NAMESPACE} get schedule.backup.appuio.ch &> /dev/null; then
        if kubectl --insecure-skip-tls-verify -n ${NAMESPACE} get schedules.backup.appuio.ch k8up-lagoon-backup-schedule &> /dev/null; then
          echo "Backups: removing old backup.appuio.ch/v1alpha1 schedule"
          kubectl --insecure-skip-tls-verify -n ${NAMESPACE} delete schedules.backup.appuio.ch k8up-lagoon-backup-schedule
        fi
        if kubectl --insecure-skip-tls-verify -n ${NAMESPACE} get prebackuppods.backup.appuio.ch &> /dev/null; then
          echo "Backups: removing old backup.appuio.ch/v1alpha1 prebackuppods"
          kubectl --insecure-skip-tls-verify -n ${NAMESPACE} delete prebackuppods.backup.appuio.ch --all
        fi
      fi
      K8UP_VERSION="v2"
    fi
  fi
  if [[ "$K8UP_VERSION" != "v2" ]]; then
    if kubectl -n ${NAMESPACE} get schedule.backup.appuio.ch &> /dev/null; then
      echo "Backups: generating backup.appuio.ch/v1alpha1 resources"
      if ! kubectl --insecure-skip-tls-verify -n ${NAMESPACE} get secret baas-repo-pw &> /dev/null; then
        # Create baas-repo-pw secret based on the project secret
        kubectl --insecure-skip-tls-verify -n ${NAMESPACE} create secret generic baas-repo-pw --from-literal=repo-pw=$(echo -n "${PROJECT_SECRET}-BAAS-REPO-PW" | sha256sum | cut -d " " -f 1)
      fi
      build-deploy-tool template backup-schedule --version v1 --saved-templates-path ${LAGOON_BACKUP_YAML_FOLDER}
    fi
  fi
  # apply backup templates
  if [ -n "$(ls -A $LAGOON_BACKUP_YAML_FOLDER/ 2>/dev/null)" ]; then
    find $LAGOON_BACKUP_YAML_FOLDER -type f -exec cat {} \;
    kubectl apply -n ${NAMESPACE} -f $LAGOON_BACKUP_YAML_FOLDER/
  fi
else
  echo ">> Backup configurations disabled for this build"
fi

currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
patchBuildStep "${buildStartTime}" "${previousStepEnd}" "${currentStepEnd}" "${NAMESPACE}" "backupConfigurationComplete" "Backup Configuration" "false"
previousStepEnd=${currentStepEnd}
beginBuildStep "Pre-Rollout Tasks" "runningPreRolloutTasks"

##############################################
### RUN PRE-ROLLOUT tasks defined in .lagoon.yml
##############################################

if [ "${LAGOON_PREROLLOUT_DISABLED}" != "true" ]; then
    build-deploy-tool tasks pre-rollout
else
  echo "pre-rollout tasks are currently disabled LAGOON_PREROLLOUT_DISABLED is set to true"
  currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
  patchBuildStep "${buildStartTime}" "${previousStepEnd}" "${currentStepEnd}" "${NAMESPACE}" "preRolloutsCompleted" "Pre-Rollout Tasks" "false"
fi

currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
previousStepEnd=${currentStepEnd}
beginBuildStep "Deployment Templating" "templatingDeployments"

##############################################
### CREATE PVC, DEPLOYMENTS AND CRONJOBS
##############################################

# generate a map of servicename>imagename+hash json for the build-deploy-tool to use when templating
# this reduces the need for the crazy logic with how services are currently mapped together in the case of nginx-php type deploymentss
touch /kubectl-build-deploy/images.yaml
for COMPOSE_SERVICE in "${COMPOSE_SERVICES[@]}"
do
  SERVICE_NAME_IMAGE_HASH="${IMAGE_HASHES[${COMPOSE_SERVICE}]}"
  yq -i '.images.'$COMPOSE_SERVICE' = "'${SERVICE_NAME_IMAGE_HASH}'"' /kubectl-build-deploy/images.yaml
done

# handle dynamic secret collection here, @TODO this will go into the state collector eventually
export DYNAMIC_SECRETS=$(kubectl -n ${NAMESPACE} get secrets -l lagoon.sh/dynamic-secret -o json | jq -r '[.items[] | .metadata.name] | join(",")')

# label subject to change
export DYNAMIC_DBAAS_SECRETS=$(kubectl -n ${NAMESPACE} get secrets -l secret.lagoon.sh/dbaas=true -o json | jq -r '[.items[] | .metadata.name] | join(",")')

# delete any custom private registry secrets, they will get re-created by the lagoon-services templates
EXISTING_REGISTRY_SECRETS=$(kubectl -n ${NAMESPACE} get secrets --no-headers | cut -d " " -f 1 | xargs)
for EXISTING_REGISTRY_SECRET in ${EXISTING_REGISTRY_SECRETS}; do
  if [[ "${EXISTING_REGISTRY_SECRET}" =~ "lagoon-private-registry-" ]]; then
    if kubectl -n ${NAMESPACE} get secret ${EXISTING_REGISTRY_SECRET} &> /dev/null; then
      kubectl -n ${NAMESPACE} delete secret ${EXISTING_REGISTRY_SECRET}
    fi
  fi
done

echo "=== BEGIN deployment template for services ==="
LAGOON_SERVICES_YAML_FOLDER="/kubectl-build-deploy/lagoon/service-deployments"
mkdir -p $LAGOON_SERVICES_YAML_FOLDER
build-deploy-tool template lagoon-services --saved-templates-path ${LAGOON_SERVICES_YAML_FOLDER} --images /kubectl-build-deploy/images.yaml

currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
patchBuildStep "${buildStartTime}" "${previousStepEnd}" "${currentStepEnd}" "${NAMESPACE}" "deploymentTemplatingComplete" "Deployment Templating" "false"
previousStepEnd=${currentStepEnd}
beginBuildStep "Applying Deployments" "applyingDeployments"

##############################################
### APPLY RESOURCES
##############################################

# remove any storage calculator pods before applying deployments to prevent storage binding issues
STORAGE_CALCULATOR_PODS=$(kubectl -n ${NAMESPACE} get pods -l lagoon.sh/storageCalculator=true --no-headers | cut -d " " -f 1 | xargs)
for STORAGE_CALCULATOR_POD in $STORAGE_CALCULATOR_PODS; do
  kubectl -n ${NAMESPACE} delete pod ${STORAGE_CALCULATOR_POD}
done

if [ "$(ls -A $LAGOON_SERVICES_YAML_FOLDER/)" ]; then
  echo "=== deployment templates for services ==="
  ls -A $LAGOON_SERVICES_YAML_FOLDER

  # cat $LAGOON_SERVICES_YAML_FOLDER/services.yaml
  # cat $LAGOON_SERVICES_YAML_FOLDER/pvcs.yaml
  # cat $LAGOON_SERVICES_YAML_FOLDER/deployments.yaml
  # cat $LAGOON_SERVICES_YAML_FOLDER/cronjobs.yaml
  if [ -n "$(ls -A $LAGOON_SERVICES_YAML_FOLDER/ 2>/dev/null)" ]; then
    find $LAGOON_SERVICES_YAML_FOLDER -type f -exec cat {} \;
    kubectl apply -n ${NAMESPACE} -f $LAGOON_SERVICES_YAML_FOLDER/
  fi
fi

##############################################
### WAIT FOR POST-ROLLOUT TO BE FINISHED
##############################################

for SERVICE_TYPES_ENTRY in "${SERVICE_TYPES[@]}"
do

  IFS=':' read -ra SERVICE_TYPES_ENTRY_SPLIT <<< "$SERVICE_TYPES_ENTRY"

  SERVICE_NAME=${SERVICE_TYPES_ENTRY_SPLIT[0]}
  SERVICE_TYPE=${SERVICE_TYPES_ENTRY_SPLIT[1]}

  # check if this service is a dbaas service and transform the service_type accordingly
  for DBAAS_ENTRY in "${DBAAS[@]}"
  do
    IFS=':' read -ra DBAAS_ENTRY_SPLIT <<< "$DBAAS_ENTRY"
    DB_SERVICE_NAME=${DBAAS_ENTRY_SPLIT[0]}
    DB_SERVICE_TYPE=${DBAAS_ENTRY_SPLIT[1]}
    if [ $SERVICE_NAME == $DB_SERVICE_NAME ]; then
      SERVICE_TYPE=$DB_SERVICE_TYPE
    fi
  done

  if [[ $SERVICE_TYPE == *"-dbaas" ]]; then
    echo "nothing to monitor for $SERVICE_TYPE"
  else
    . /kubectl-build-deploy/scripts/exec-monitor-deploy.sh
  fi
done

if kubectl -n ${NAMESPACE} get configmap lagoon-env &> /dev/null; then
  # now delete the configmap after all the lagoon-env and lagoon-platform-env calcs have been done
  # and the deployments have rolled out successfully, this makes less problems rolling back if a build fails
  # somewhere between the new secret being created, and the deployments rolling out
  kubectl -n ${NAMESPACE} delete configmap lagoon-env
fi

currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
patchBuildStep "${buildStartTime}" "${previousStepEnd}" "${currentStepEnd}" "${NAMESPACE}" "deploymentApplyComplete" "Applying Deployments" "false"
previousStepEnd=${currentStepEnd}
beginBuildStep "Cronjob Cleanup" "cleaningUpCronjobs"

##############################################
### CLEANUP NATIVE CRONJOBS which have been removed from .lagoon.yml or modified to run more frequently than every 15 minutes
##############################################

CURRENT_CRONJOBS=$(kubectl -n ${NAMESPACE} get cronjobs --no-headers | cut -d " " -f 1 | xargs)
MATCHED_CRONJOB=false
DELETE_CRONJOBS=()
NATIVE_CRONJOB_CLEANUP_ARRAY=$(build-deploy-tool identify native-cronjobs | jq -r '.[]')
for SINGLE_NATIVE_CRONJOB in $CURRENT_CRONJOBS; do
  for CLEANUP_NATIVE_CRONJOB in ${NATIVE_CRONJOB_CLEANUP_ARRAY[@]}; do
    if [ "${SINGLE_NATIVE_CRONJOB}" == "${CLEANUP_NATIVE_CRONJOB}" ]; then
      MATCHED_CRONJOB=true
      continue
    fi
  done
  if [ "${MATCHED_CRONJOB}" != "true" ]; then
    DELETE_CRONJOBS+=($SINGLE_NATIVE_CRONJOB)
  fi
  MATCHED_CRONJOB=false
done
for DC in ${!DELETE_CRONJOBS[@]}; do
  # delete any cronjobs if they were removed
  if kubectl -n ${NAMESPACE} get cronjob ${DELETE_CRONJOBS[$DC]} &> /dev/null; then
    echo ">> Removing cronjob ${DELETE_CRONJOBS[$DC]} because it was removed"
    kubectl -n ${NAMESPACE} delete cronjob ${DELETE_CRONJOBS[$DC]}
  fi
done

currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
patchBuildStep "${buildStartTime}" "${previousStepEnd}" "${currentStepEnd}" "${NAMESPACE}" "cronjobCleanupComplete" "Cronjob Cleanup" "false"
previousStepEnd=${currentStepEnd}
beginBuildStep "Post-Rollout Tasks" "runningPostRolloutTasks"

##############################################
### RUN POST-ROLLOUT tasks defined in .lagoon.yml
##############################################

# if we have LAGOON_POSTROLLOUT_DISABLED set, don't try to run any pre-rollout tasks
if [ "${LAGOON_POSTROLLOUT_DISABLED}" != "true" ]; then
  build-deploy-tool tasks post-rollout
else
  echo "post-rollout tasks are currently disabled LAGOON_POSTROLLOUT_DISABLED is set to true"
  currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
  patchBuildStep "${buildStartTime}" "${previousStepEnd}" "${currentStepEnd}" "${NAMESPACE}" "postRolloutsCompleted" "Post-Rollout Tasks" "false"
fi

currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
previousStepEnd=${currentStepEnd}
beginBuildStep "Build and Deploy" "finalizingBuild"

##############################################
### PUSH the latest .lagoon.yml into lagoon-yaml configmap
##############################################

echo "Updating lagoon-yaml configmap with a post-deploy version of the .lagoon.yml file"
if kubectl -n ${NAMESPACE} get configmap lagoon-yaml &> /dev/null; then
  # replace it, no need to check if the key is different, as that will happen in the pre-deploy phase
  kubectl -n ${NAMESPACE} get configmap lagoon-yaml -o json | jq --arg add "`cat .lagoon.yml`" '.data."post-deploy" = $add' | kubectl apply -f -
 else
  # create it
  kubectl -n ${NAMESPACE} create configmap lagoon-yaml --from-file=post-deploy=.lagoon.yml
fi
echo "Updating docker-compose-yaml configmap with a post-deploy version of the docker-compose.yml file"
if kubectl -n ${NAMESPACE} get configmap docker-compose-yaml &> /dev/null; then
  # replace it, no need to check if the key is different, as that will happen in the pre-deploy phase
  kubectl -n ${NAMESPACE} get configmap docker-compose-yaml -o json | jq --arg add "`cat ${DOCKER_COMPOSE_YAML}`" '.data."post-deploy" = $add' | kubectl apply -f -
 else
  # create it
  kubectl -n ${NAMESPACE} create configmap docker-compose-yaml --from-file=post-deploy=${DOCKER_COMPOSE_YAML}
fi

# remove any certificates for tls-acme false ingress to prevent reissuing attempts
TLS_FALSE_INGRESSES=$(kubectl -n ${NAMESPACE} get ingress -o json | jq -r '.items[] | select(.metadata.annotations["kubernetes.io/tls-acme"] == "false") | .metadata.name')
for TLS_FALSE_INGRESS in $TLS_FALSE_INGRESSES; do
  TLS_SECRETS=$(kubectl -n ${NAMESPACE} get ingress ${TLS_FALSE_INGRESS} -o json | jq -r '.spec.tls[]?.secretName')
  for TLS_SECRET in $TLS_SECRETS; do
    echo ">> Cleaning up certificate for ${TLS_SECRET} as tls-acme is set to false"
    # check if it is a lets encrypt certificate
    if kubectl -n ${NAMESPACE} get secret ${TLS_SECRET} &> /dev/null; then
      if openssl x509 -in <(kubectl -n ${NAMESPACE} get secret ${TLS_SECRET} -o json | jq -r '.data."tls.crt" | @base64d') -text -noout | grep -o -q "Let's Encrypt" &> /dev/null; then
        kubectl -n ${NAMESPACE} delete secret ${TLS_SECRET}
      fi
    fi
    if kubectl -n ${NAMESPACE} get certificates.cert-manager.io ${TLS_SECRET} &> /dev/null; then
      kubectl -n ${NAMESPACE} delete certificates.cert-manager.io ${TLS_SECRET}
    fi
  done
done

currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
patchBuildStep "${buildStartTime}" "${previousStepEnd}" "${currentStepEnd}" "${NAMESPACE}" "deployCompleted" "Build and Deploy" "false"
previousStepEnd=${currentStepEnd}

if [ "$(featureFlag INSIGHTS)" = enabled ]; then
  beginBuildStep "Insights Gathering" "gatheringInsights"
  ##############################################
  ### RUN insights gathering and store in configmap
  ##############################################
  INSIGHTS_WARNING_COUNT=0
  for IMAGE_NAME in "${!IMAGES_BUILD[@]}"
  do
    IMAGE_TAG="${IMAGE_TAG:-latest}"
    IMAGE_FULL="${REGISTRY}/${PROJECT}/${ENVIRONMENT}/${IMAGE_NAME}:${IMAGE_TAG}"
    insightsOutput=$(. /kubectl-build-deploy/scripts/exec-generate-insights-configmap.sh)
    if (exit $?); then
      echo "${insightsOutput}"
    else
      ((++INSIGHTS_WARNING_COUNT))
      echo "> This insights run failed, this warning is for information only."
      echo "${insightsOutput}"
    fi
  done
  if [[ "$INSIGHTS_WARNING_COUNT" -gt 0 ]]; then
    ((++BUILD_WARNING_COUNT))
    echo "##############################################"
    currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
    patchBuildStep "${buildStartTime}" "${previousStepEnd}" "${currentStepEnd}" "${NAMESPACE}" "insightsWarning" "Insights Gathering" "true"
    previousStepEnd=${currentStepEnd}
  else
    currentStepEnd="$(date +"%Y-%m-%d %H:%M:%S")"
    patchBuildStep "${buildStartTime}" "${previousStepEnd}" "${currentStepEnd}" "${NAMESPACE}" "insightsCompleted" "Insights Gathering" "false"
    previousStepEnd=${currentStepEnd}
  fi

fi

if [[ "$BUILD_WARNING_COUNT" -gt 0 ]]; then
  beginBuildStep "Completed With Warnings" "deployCompletedWithWarnings"
  echo "This build completed with ${BUILD_WARNING_COUNT} warnings, you should scan the build for warnings and correct them as neccessary"
  patchBuildStep "${buildStartTime}" "${previousStepEnd}" "${currentStepEnd}" "${NAMESPACE}" "deployCompletedWithWarnings" "Completed With Warnings" "true"
  previousStepEnd=${currentStepEnd}
  # patch the buildpod with the buildstep
  if [ "${SCC_CHECK}" == false ]; then
    kubectl patch -n ${NAMESPACE} pod ${LAGOON_BUILD_NAME} \
      -p "{\"metadata\":{\"labels\":{\"lagoon.sh/buildStep\":\"deployCompletedWithWarnings\"}}}" &> /dev/null
    # tiny sleep to allow patch to complete before logs roll again
    sleep 5
  fi
fi
