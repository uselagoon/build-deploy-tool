#!/bin/bash

# Handle GPU device requests
GPU_REQUEST=$(cat $DOCKER_COMPOSE_YAML | shyaml get-value services.$COMPOSE_SERVICE.labels.lagoon\\.gpu false)
if [ ! $GPU_REQUEST == "false" ]; then
  GPU_REQUEST_SIZE=$(cat $DOCKER_COMPOSE_YAML | shyaml get-value services.$COMPOSE_SERVICE.labels.lagoon\\.gpu\\.size false)
  if [ ! $GPU_REQUEST_SIZE == "false" ]; then
    GPU_SIZE=$GPU_REQUEST_SIZE
  else
    GPU_SIZE=1
  fi

  echo -e "\
tolerations:
- key: lagoon.sh/gpu
  operator: Equal
  value: 'true'
  effect: NoSchedule
nodeSelector:
  lagoon.sh/gpu: 'true'
resources:
  limits:
    nvidia.com/gpu: ${GPU_SIZE}
" >> /kubectl-build-deploy/${SERVICE_NAME}-values.yaml
fi
