#!/bin/bash


if [[ -z "$DOCKER_COMPOSE_YAML" ]]; then
  echo "no docker compose file given"
  exit
fi

EXTRA_MOUNT_VALUES_FILE="${KBD_SERVICE_VALUES_OUTDIR:-kubectl-build-deploy}/extravolumes-values.yaml"

# The prefix below CAN be used, but this should be moved into the helm templates
CUSTOMVOLUME_PREFIX="" # we just use this to distinguish from any other volumes that might be created

# Parse docker-compose.yml and extract volume names with "lagoon.type: persistent" label
volumes=$(yq e '.volumes | with_entries(select(.value.labels."lagoon.type" == "persistent")) | keys | .[]' "$DOCKER_COMPOSE_YAML")

# Print the list of volume names
echo "Extra volumes defined:"
echo "$volumes"
echo

# Create an array to store the volumes that need to be created
volumes_to_create=()
EXTRA_VOLUMES_MOUNT_VALS="\
customVolumeMounts:
" #this will be output to our values file

# Iterate over the volumes
for volume in $volumes; do
#  echo "Volume: $volume"
  EXTRA_VOLUMES_MOUNT_VALS+="\
  - $CUSTOMVOLUME_PREFIX$volume:
"
  # Loop through the services and check if they reference the current volume
  services=$(yq e '.services | to_entries | .[] | select(.value.labels | has("lagoon.volumes.'$volume'.path")) | .key' "$DOCKER_COMPOSE_YAML")

  # Print the services and their corresponding paths for the current volume
  while IFS= read -r service; do
    path=$(yq e '.services."'$service'".labels."lagoon.volumes.'$volume'.path"' "$DOCKER_COMPOSE_YAML")
#    echo "- Service: $service, Path: $path"

  if [[ "$service" != "" ]]; then
    EXTRA_VOLUMES_MOUNT_VALS+="\
    - $service: $path
"
  fi

  done <<< "$services"

  # If no services reference the volume, print a message indicating it is not used
  if [[ -z "$services" ]]; then
    echo "- Not used"
  else
    # Add the volume to the array of volumes that need to be created
    volumes_to_create+=("$volume")
  fi

  echo
done

echo "Volumes to be created:"
echo "${volumes_to_create[@]}"

EXTRA_VOLUMES_VALUES_YAML=""


# Check if volumes_to_create array is not empty before iterating
if [[ ${#volumes_to_create[@]} -gt 0 ]]; then
  echo "Volumes to be created:"
  EXTRA_VOLUMES_VALUES_YAML+="\
customVolumes:
"
  for volume in "${volumes_to_create[@]}"; do
    echo "- $CUSTOMVOLUME_PREFIX$volume"
    EXTRA_VOLUMES_VALUES_YAML+="\
    - $volume
"
  done
else
  echo "No volumes to create."
fi

echo "$EXTRA_VOLUMES_VALUES_YAML"
echo "$EXTRA_VOLUMES_MOUNT_VALS"

echo "$EXTRA_VOLUMES_VALUES_YAML" > $EXTRA_MOUNT_VALUES_FILE
echo "$EXTRA_VOLUMES_MOUNT_VALS" >> $EXTRA_MOUNT_VALUES_FILE