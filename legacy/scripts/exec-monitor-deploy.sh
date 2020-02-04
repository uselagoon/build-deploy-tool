#!/bin/bash

# while the rollout of a new deployment is running we gather the logs of the new generated pods and save them in a known location
# in case this rollout fails, we show the logs of the new containers to the user as they might contain information about why
# the rollout has failed
stream_logs_deployment() {
  set +x
  # load the version of the new pods
  LATEST_POD_TEMPLATE_HASH=$(kubectl get replicaset -l app.kubernetes.io/instance=${SERVICE_NAME} --sort-by=.metadata.creationTimestamp -o=json | jq -r '.items[-1].metadata.labels."pod-template-hash"')
  mkdir -p /tmp/kubectl-build-deploy/logs/container/${SERVICE_NAME}

  # this runs in a loop forever (until killed)
  while [ 1 ]
  do
    # Gatter all pods and their containers for the current rollout and stream their logs into files
    kubectl -n ${NAMESPACE} get --insecure-skip-tls-verify pods -l pod-template-hash=${LATEST_POD_TEMPLATE_HASH} -o json | jq -r '.items[] | .metadata.name + " " + .spec.containers[].name' |
    {
      while read -r POD CONTAINER ; do
          kubectl -n ${NAMESPACE} logs --insecure-skip-tls-verify --timestamps -f $POD -c $CONTAINER $SINCE_TIME 2> /dev/null > /tmp/oc-build-deploy/logs/container/${SERVICE_NAME}/$POD-$CONTAINER.log &
      done

      # this will wait for all log streaming we started to finish
      wait
    }

    # If we are here, this means the pods have all stopped (probably because they failed), we just restart
  done
}

# start background logs streaming
stream_logs_deployment &
STREAM_LOGS_PID=$!

ret=0
kubectl rollout --insecure-skip-tls-verify -n ${NAMESPACE} status deployment ${SERVICE_NAME} --watch || ret=$?

if [[ $ret -ne 0 ]]; then
  # stop all running stream logs
  pkill -P $STREAM_LOGS_PID || true

  # shows all logs we collected for the new containers
  if [ -z "$(ls -A /tmp/kubectl-build-deploy/logs/container/${SERVICE_NAME})" ]; then
    echo "Rollout for ${SERVICE_NAME} failed, tried to gather some startup logs of the containers, but unfortunately there were none created, sorry."
  else
    echo "Rollout for ${SERVICE_NAME} failed, tried to gather some startup logs of the containers, hope this helps debugging:"
    find /tmp/kubectl-build-deploy/logs/container/${SERVICE_NAME}/ -type f -print0 2>/dev/null | xargs -0 -I % sh -c 'echo ======== % =========; cat %; echo'
  fi

  exit 1
fi

# stop all running stream logs
pkill -P $STREAM_LOGS_PID || true
