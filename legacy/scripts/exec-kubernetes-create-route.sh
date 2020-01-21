#!/bin/bash

# TODO: find out why we are using the if/else and if it's still needed for kubernetes
if oc --insecure-skip-tls-verify -n ${OPENSHIFT_PROJECT} get route "$ROUTE_DOMAIN" &> /dev/null; then
  oc --insecure-skip-tls-verify -n ${OPENSHIFT_PROJECT} patch route "$ROUTE_DOMAIN" -p "{\"metadata\":{\"annotations\":{\"kubernetes.io/tls-acme\":\"${ROUTE_TLS_ACME}\",\"haproxy.router.openshift.io/hsts_header\":\"${ROUTE_HSTS}\"}},\"spec\":{\"to\":{\"name\":\"${ROUTE_SERVICE}\"},\"tls\":{\"insecureEdgeTerminationPolicy\":\"${ROUTE_INSECURE}\"}}}"
else
  oc process  --local -o yaml --insecure-skip-tls-verify \
    -n ${OPENSHIFT_PROJECT} \
    -f /oc-build-deploy/openshift-templates/route.yml \
    -p SAFE_BRANCH="${SAFE_BRANCH}" \
    -p SAFE_PROJECT="${SAFE_PROJECT}" \
    -p BRANCH="${BRANCH}" \
    -p PROJECT="${PROJECT}" \
    -p LAGOON_GIT_SHA="${LAGOON_GIT_SHA}" \
    -p OPENSHIFT_PROJECT=${OPENSHIFT_PROJECT} \
    -p ROUTE_DOMAIN="${ROUTE_DOMAIN}" \
    -p ROUTE_SERVICE="${ROUTE_SERVICE}" \
    -p ROUTE_TLS_ACME="${ROUTE_TLS_ACME}" \
    -p ROUTE_INSECURE="${ROUTE_INSECURE}" \
    -p ROUTE_HSTS="${ROUTE_HSTS}" \
    | outputToYaml
fi
