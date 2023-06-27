#!/bin/bash

# First step: Parse the docker-compose.yml for any volumes that have the label "lagoon.type: persistent"
DOCKER_COMPOSE_YAML="test/docker-compose.yml"
# Parse docker-compose.yml and extract volume names with "lagoon.type: persistent" label
volumes=$(yq e '.volumes | with_entries(select(.value.labels."lagoon.type" == "persistent")) | keys | .[]' "$DOCKER_COMPOSE_YAML")

# Print the list of volume names
echo "Volumes:"
echo "$volumes"
echo

# Create an array to store the volumes that need to be created
volumes_to_create=()

# Iterate over the volumes
for volume in $volumes; do
  echo "Volume: $volume"

  # Loop through the services and check if they reference the current volume
  services=$(yq e '.services | to_entries | .[] | select(.value.labels | has("lagoon.volumes.'$volume'.path")) | .key' "$DOCKER_COMPOSE_YAML")

  # Print the services and their corresponding paths for the current volume
  while IFS= read -r service; do
    path=$(yq e '.services."'$service'".labels."lagoon.volumes.'$volume'.path"' "$DOCKER_COMPOSE_YAML")
    echo "- Service: $service, Path: $path"
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
