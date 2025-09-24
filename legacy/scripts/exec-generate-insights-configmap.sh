#!/bin/bash

TMP_DIR="${TMP_DIR:-/tmp}"
SBOM_OUTPUT="cyclonedx"

SBOM_OUTPUT_FILE="${TMP_DIR}/${IMAGE_NAME}.cyclonedx.json.gz"
SBOM_CONFIGMAP="lagoon-insights-sbom-${IMAGE_NAME}"
IMAGE_INSPECT_CONFIGMAP="lagoon-insights-image-${IMAGE_NAME}"
IMAGE_INSPECT_OUTPUT_FILE="${TMP_DIR}/${IMAGE_NAME}.image-inspect.json.gz"

# Here we give the cluster administrator the ability to override the insights scan image
INSIGHTS_SCAN_IMAGE="uselagoon/insights-trivy"
  if [ "$ADMIN_LAGOON_FEATURE_FLAG_INSIGHTS_SCAN_IMAGE" ]; then
    INSIGHTS_SCAN_IMAGE="${ADMIN_LAGOON_FEATURE_FLAG_INSIGHTS_SCAN_IMAGE}"
  fi

set +x
echo "Running image inspect on: ${IMAGE_FULL}"

skopeo inspect --retry-times 5 docker://${IMAGE_FULL} --tls-verify=false | gzip > ${IMAGE_INSPECT_OUTPUT_FILE}

processImageInspect() {
  echo "Successfully generated image inspection data for ${IMAGE_FULL}"

  # If lagoon-insights-image-inspect-[IMAGE] configmap already exists then we need to update, else create new
  if kubectl -n ${NAMESPACE} get configmap $IMAGE_INSPECT_CONFIGMAP &> /dev/null; then
      kubectl \
          -n ${NAMESPACE} \
          create configmap $IMAGE_INSPECT_CONFIGMAP \
          --from-file=${IMAGE_INSPECT_OUTPUT_FILE} \
          -o json \
          --dry-run=client | kubectl replace -f -
  else
      kubectl \
          -n ${NAMESPACE} \
          create configmap ${IMAGE_INSPECT_CONFIGMAP} \
          --from-file=${IMAGE_INSPECT_OUTPUT_FILE}
  fi
  if [[ "$BUILD_TYPE" == "pullrequest" ]]; then
    kubectl -n ${NAMESPACE} \
      annotate configmap ${IMAGE_INSPECT_CONFIGMAP} \
      lagoon.sh/branch=${PR_NUMBER} \
      lagoon.sh/prHeadBranch=${PR_HEAD_BRANCH} \
      lagoon.sh/prBaseBranch=${PR_BASE_BRANCH}
  else
    kubectl -n ${NAMESPACE} \
      annotate configmap ${IMAGE_INSPECT_CONFIGMAP} \
      lagoon.sh/branch=${BRANCH}
  fi
  kubectl \
      -n ${NAMESPACE} \
      label configmap ${IMAGE_INSPECT_CONFIGMAP} \
      lagoon.sh/insightsProcessed- \
      lagoon.sh/insightsType=image-gz \
      lagoon.sh/buildName=${LAGOON_BUILD_NAME} \
      lagoon.sh/project=${PROJECT} \
      lagoon.sh/environment=${ENVIRONMENT} \
      lagoon.sh/service=${IMAGE_NAME} \
      lagoon.sh/environmentType=${ENVIRONMENT_TYPE} \
      lagoon.sh/buildType=${BUILD_TYPE} \
      insights.lagoon.sh/type=inspect
}

processImageInspect

echo "Running sbom scan using trivy"
echo "Image being scanned: ${IMAGE_FULL}"
echo "Using image for scan ${IMAGECACHE_REGISTRY}${INSIGHTS_SCAN_IMAGE}"

# Setting JAVAOPT to skip the java db update, as the upstream image comes with a pre-populated database
JAVAOPT="--skip-java-db-update"
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock ${IMAGECACHE_REGISTRY}${INSIGHTS_SCAN_IMAGE} image ${JAVAOPT} ${IMAGE_FULL} --format ${SBOM_OUTPUT} --skip-version-check | gzip > ${SBOM_OUTPUT_FILE}

FILESIZE=$(stat -c%s "$SBOM_OUTPUT_FILE")
echo "Size of ${SBOM_OUTPUT_FILE} = $FILESIZE bytes."

processSbom() {
  if (( $FILESIZE > 950000 )); then
    echo "$SBOM_OUTPUT_FILE is too large, skipping pushing to configmap"
    return
  else
    echo "Successfully generated SBOM for ${IMAGE_FULL}"

    # If lagoon-insights-sbom-[IMAGE] configmap already exists then we need to update, else create new
    if kubectl -n ${NAMESPACE} get configmap $SBOM_CONFIGMAP &> /dev/null; then
        kubectl \
            -n ${NAMESPACE} \
            create configmap $SBOM_CONFIGMAP \
            --from-file=${SBOM_OUTPUT_FILE} \
            -o json \
            --dry-run=client | kubectl replace -f -
    else
        # Create configmap and add label (#have to add label separately: https://github.com/kubernetes/kubernetes/issues/60295)
        kubectl \
            -n ${NAMESPACE} \
            create configmap ${SBOM_CONFIGMAP} \
            --from-file=${SBOM_OUTPUT_FILE}
    fi
    if [[ "$BUILD_TYPE" == "pullrequest" ]]; then
      kubectl -n ${NAMESPACE} \
        annotate configmap ${SBOM_CONFIGMAP} \
        lagoon.sh/branch=${PR_NUMBER} \
        lagoon.sh/prHeadBranch=${PR_HEAD_BRANCH} \
        lagoon.sh/prBaseBranch=${PR_BASE_BRANCH}
    else
      kubectl -n ${NAMESPACE} \
        annotate configmap ${SBOM_CONFIGMAP} \
        lagoon.sh/branch=${BRANCH}
    fi
    # Support custom Dependency Track integration.
    local apiEndpoint
    apiEndpoint=$(featureFlag INSIGHTS_DEPENDENCY_TRACK_API_ENDPOINT)
    local apiKey
    apiKey=$(featureFlag INSIGHTS_DEPENDENCY_TRACK_API_KEY)
    local dtWarn
    if [ -n "$apiEndpoint" ]; then
      if [ -n "$apiKey" ]; then
        # Test API access
        local resp
        if ! resp=$(curl -sSf -m 60 -H "X-Api-Key:${apiKey}" "${apiEndpoint}/api/v1/project?pageSize=1" 2>&1); then
          dtWarn="\n\n**********\nCustom Dependency Track not enabled: API Error: ${resp}\n**********\n\n"
        else
          kubectl -n ${NAMESPACE} \
            annotate configmap ${SBOM_CONFIGMAP} \
            dependencytrack.insights.lagoon.sh/custom-endpoint="${apiEndpoint}"
        fi
      else
        dtWarn="\n\n**********\nCustom Dependency Track not enabled: Missing LAGOON_FEATURE_FLAG_INSIGHTS_DEPENDENCY_TRACK_API_KEY\n**********\n\n"
      fi
    fi
    kubectl \
        -n ${NAMESPACE} \
        label configmap ${SBOM_CONFIGMAP} \
        lagoon.sh/insightsProcessed- \
        lagoon.sh/insightsType=sbom-gz \
        lagoon.sh/buildName=${LAGOON_BUILD_NAME} \
        lagoon.sh/project=${PROJECT} \
        lagoon.sh/environment=${ENVIRONMENT} \
        lagoon.sh/service=${IMAGE_NAME} \
        lagoon.sh/environmentType=${ENVIRONMENT_TYPE} \
        lagoon.sh/buildType=${BUILD_TYPE} \
        insights.lagoon.sh/type=sbom

    if [ -n "$dtWarn" ]; then
      printf '%b' "$dtWarn"
      return 1
    fi
  fi
}

processSbom
