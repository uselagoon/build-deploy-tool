#!/bin/bash

# Dynamic secret loading
# This script will look in the current namespace for any secrets that have been

DYNAMIC_SECRET_LABEL="lagoon.sh/dynamic-secret"

KBD_SERVICE_VALUES_FILE="/${KBD_SERVICE_VALUES_OUTDIR:-kubectl-build-deploy}/${SERVICE_NAME}-values.yaml"

VOLUME_MOUNT_BASE_PATH="/var/run/secrets/lagoon/dynamic/"

VOLUME_NAME_PREFIX="dynamic-"
SECRET_NAME_PREFIX="dynamic-"

RAW_KUBECTL_JSON_SECRET_LIST=$(kubectl --namespace ${NAMESPACE} get secrets -l $DYNAMIC_SECRET_LABEL -o json)


SECRET_MOUNT_VALUES=$'dynamicSecretMounts:\n'
SECRET_VOL_VALUES=$'dynamicSecretVolumes:\n'

echo "$RAW_KUBECTL_JSON_SECRET_LIST" | jq -c --raw-output '.items[] | .metadata.name' | (
  while IFS=$"\n" read -r name; do
    # so we have to do two things here. Generate the volume and the mount
    MOUNT_PATH="$VOLUME_MOUNT_BASE_PATH$name"
    SECRET_NAME="$name"
    VOLUME_NAME="$VOLUME_NAME_PREFIX$name"
    SECRET_MOUNT_VALUES+="\
  - name: $VOLUME_NAME
    mountPath: "$MOUNT_PATH"
    readOnly: true
"

    SECRET_VOL_VALUES+="\
  - name: $VOLUME_NAME
    secret:
      secretName: $SECRET_NAME
      optional: false
"
  done
  echo "$SECRET_MOUNT_VALUES" >> $KBD_SERVICE_VALUES_FILE
  echo "$SECRET_VOL_VALUES" >> $KBD_SERVICE_VALUES_FILE
)
