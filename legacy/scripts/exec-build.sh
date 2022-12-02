#!/bin/bash

# try to pull the last pushed image so we can use it for --cache-from during the build
set +x

if [ $BUILD_TARGET == "false" ]; then
    echo "Building ${BUILD_CONTEXT}/${DOCKERFILE}"
    DOCKER_BUILDKIT=$DOCKER_BUILDKIT docker build --network=host "${BUILD_ARGS[@]}" -t $TEMPORARY_IMAGE_NAME -f $BUILD_CONTEXT/$DOCKERFILE $BUILD_CONTEXT
else
    echo "Building target ${BUILD_TARGET} for ${BUILD_CONTEXT}/${DOCKERFILE}"
    DOCKER_BUILDKIT=$DOCKER_BUILDKIT docker build --network=host "${BUILD_ARGS[@]}" -t $TEMPORARY_IMAGE_NAME -f $BUILD_CONTEXT/$DOCKERFILE --target $BUILD_TARGET $BUILD_CONTEXT
fi
set -x
