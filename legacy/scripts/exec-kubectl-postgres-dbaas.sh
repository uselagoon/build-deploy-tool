#!/bin/bash

# The operator can sometimes take a bit, wait until the details are available
# We added a timeout of 5 minutes (60 retries) before exit
OPERATOR_COUNTER=1
OPERATOR_TIMEOUT=60
# use the secret name from the consumer to prevent credential clash
until kubectl -n ${NAMESPACE} get postgresqlconsumer/${SERVICE_NAME} -o yaml | shyaml get-value spec.consumer.database
do
if [ $OPERATOR_COUNTER -lt $OPERATOR_TIMEOUT ]; then
    consumer_failed=$(kubectl -n ${NAMESPACE} get postgresqlconsumer/${SERVICE_NAME} -o json | jq -r '.metadata.annotations."dbaas.amazee.io/failed"')
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

# Grab the details from the consumer spec
DB_HOST=$(kubectl -n ${NAMESPACE} get postgresqlconsumer/${SERVICE_NAME} -o yaml | shyaml get-value spec.consumer.services.primary)
DB_USER=$(kubectl -n ${NAMESPACE} get postgresqlconsumer/${SERVICE_NAME} -o yaml | shyaml get-value spec.consumer.username)
DB_PASSWORD=$(kubectl -n ${NAMESPACE} get postgresqlconsumer/${SERVICE_NAME} -o yaml | shyaml get-value spec.consumer.password)
DB_NAME=$(kubectl -n ${NAMESPACE} get postgresqlconsumer/${SERVICE_NAME} -o yaml | shyaml get-value spec.consumer.database)
DB_PORT=$(kubectl -n ${NAMESPACE} get postgresqlconsumer/${SERVICE_NAME} -o yaml | shyaml get-value spec.provider.port)

# Add credentials to our configmap, prefixed with the name of the servicename of this servicebroker
kubectl patch \
    -n ${NAMESPACE} \
    configmap lagoon-env \
    -p "{\"data\":{\"${SERVICE_NAME_UPPERCASE}_HOST\":\"${DB_HOST}\", \"${SERVICE_NAME_UPPERCASE}_USERNAME\":\"${DB_USER}\", \"${SERVICE_NAME_UPPERCASE}_PASSWORD\":\"${DB_PASSWORD}\", \"${SERVICE_NAME_UPPERCASE}_DATABASE\":\"${DB_NAME}\", \"${SERVICE_NAME_UPPERCASE}_PORT\":\"${DB_PORT}\"}}"

# only add the DB_READREPLICA_HOSTS variable if it exists in the consumer spec
# since the operator can support multiple replica hosts being defined, we should comma seperate them here
if DB_READREPLICA_HOSTS=$(kubectl -n ${NAMESPACE} get postgresqlconsumer/${SERVICE_NAME} -o yaml | shyaml get-value spec.consumer.services.replicas); then
    DB_READREPLICA_HOSTS=$(echo $DB_READREPLICA_HOSTS | cut -c 3- | rev | cut -c 1- | rev | sed 's/^\|$//g' | paste -sd, -)
    yq3 write -i -- /kubectl-build-deploy/${SERVICE_NAME}-values.yaml 'readReplicaHosts' $DB_READREPLICA_HOSTS
    kubectl patch \
        -n ${NAMESPACE} \
        configmap lagoon-env \
        -p "{\"data\":{\"${SERVICE_NAME_UPPERCASE}_READREPLICA_HOSTS\":\"${DB_READREPLICA_HOSTS}\"}}"
fi
