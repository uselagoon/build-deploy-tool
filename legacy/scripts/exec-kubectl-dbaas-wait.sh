#!/bin/bash

# The operator can sometimes take a bit, wait until the details are available
# We added a timeout of 5 minutes (60 retries) before exit
OPERATOR_COUNTER=1
OPERATOR_TIMEOUT=60
# use the secret name from the consumer to prevent credential clash
until [ "$(kubectl -n ${NAMESPACE} get ${CONSUMER_TYPE}/${SERVICE_NAME} -o json | jq -r '.spec.consumer.database')" != "null" ];
do
if [ $OPERATOR_COUNTER -lt $OPERATOR_TIMEOUT ]; then
    consumer_failed=$(kubectl -n ${NAMESPACE} get ${CONSUMER_TYPE}/${SERVICE_NAME} -o json | jq -r '.metadata.annotations."dbaas.amazee.io/failed"')
    if [ "${consumer_failed}" == "true" ]; then
        echo "Failed to provision a database. Contact your support team to investigate."
        exit 1
    fi
    let OPERATOR_COUNTER=OPERATOR_COUNTER+1
    echo "Service for ${SERVICE_NAME} not available yet, waiting for 5 secs"
    sleep 5
else
    echo "Timeout of $OPERATOR_TIMEOUT for ${SERVICE_NAME} creation reached"
    exit 1
fi
done

