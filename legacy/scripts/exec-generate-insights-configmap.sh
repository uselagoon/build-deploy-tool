#!/bin/bash

TMP_DIR="${TMP_DIR:-/tmp}"
SBOM_OUTPUT="cyclonedx"

SBOM_OUTPUT_FILE="${TMP_DIR}/${IMAGE_NAME}.cyclonedx.json.gz"
SBOM_CONFIGMAP="lagoon-insights-sbom-${IMAGE_NAME}"
IMAGE_INSPECT_CONFIGMAP="lagoon-insights-image-${IMAGE_NAME}"
IMAGE_INSPECT_OUTPUT_FILE="${TMP_DIR}/${IMAGE_NAME}.image-inspect.json.gz"

# Here we give the cluster administrator the ability to override the insights scan image
INSIGHTS_SCAN_IMAGE="aquasec/trivy"
  if [ "$ADMIN_LAGOON_FEATURE_FLAG_INSIGHTS_SCAN_IMAGE" ]; then
    INSIGHTS_SCAN_IMAGE="${ADMIN_LAGOON_FEATURE_FLAG_INSIGHTS_SCAN_IMAGE}"
  fi

set +x
echo "Running image inspect on: ${IMAGE_FULL}"

skopeo inspect --retry-times 5 docker://${IMAGE_FULL} --tls-verify=false | gzip > ${IMAGE_INSPECT_OUTPUT_FILE}

processImageInspect() {
  echo "Successfully generated image inspection data for ${IMAGE_FULL}"

  # If lagoon-insights-image-inpsect-[IMAGE] configmap already exists then we need to update, else create new
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
  kubectl \
      -n ${NAMESPACE} \
      label configmap ${IMAGE_INSPECT_CONFIGMAP} \
      lagoon.sh/insightsProcessed- \
      lagoon.sh/insightsType=image-gz \
      lagoon.sh/buildName=${LAGOON_BUILD_NAME} \
      lagoon.sh/project=${PROJECT} \
      lagoon.sh/environment=${ENVIRONMENT} \
      lagoon.sh/service=${IMAGE_NAME} \
      insights.lagoon.sh/type=inspect
}

processImageInspect

echo "Running sbom scan using trivy"
echo "Image being scanned: ${IMAGE_FULL}"
echo "Using image for scan ${IMAGECACHE_REGISTRY}${INSIGHTS_SCAN_IMAGE}"

DOCKER_HOST=docker-host.lagoon.svc docker run --rm -v /var/run/docker.sock:/var/run/docker.sock ${IMAGECACHE_REGISTRY}${INSIGHTS_SCAN_IMAGE} image --skip-java-db-update ${IMAGE_FULL} --format ${SBOM_OUTPUT} | gzip > ${SBOM_OUTPUT_FILE}

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
    kubectl \
        -n ${NAMESPACE} \
        label configmap ${SBOM_CONFIGMAP} \
        lagoon.sh/insightsProcessed- \
        lagoon.sh/insightsType=sbom-gz \
        lagoon.sh/buildName=${LAGOON_BUILD_NAME} \
        lagoon.sh/project=${PROJECT} \
        lagoon.sh/environment=${ENVIRONMENT} \
        lagoon.sh/service=${IMAGE_NAME} \
        insights.lagoon.sh/type=sbom
  fi
}

processSbom
