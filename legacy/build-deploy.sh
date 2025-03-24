#!/bin/bash
set -eo pipefail
set -o noglob

# print out the build-deploy-tool version information
echo "##############################################"
build-deploy-tool version
echo "##############################################"

NAMESPACE=$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace)
LAGOON_VERSION=$(cat /lagoon/version)
export NAMESPACE
export LAGOON_VERSION

echo -e "##############################################\nBEGIN Checkout Repository\n##############################################"
# check if a http/https url is defined, and if a username/password are supplied for it
if [ "$(build-deploy-tool template git-credential --credential-file /home/.git-credentials)" == "store" ]; then
  git config --global credential.helper 'store --file /home/.git-credentials'
fi
if [ "$BUILD_TYPE" == "pullrequest" ]; then
  /kubectl-build-deploy/scripts/git-checkout-pull-merge.sh "$SOURCE_REPOSITORY" "$PR_HEAD_SHA" "$PR_BASE_SHA"
else
  /kubectl-build-deploy/scripts/git-checkout-pull.sh "$SOURCE_REPOSITORY" "$GIT_REF"
fi

if [[ -n "$SUBFOLDER" ]]; then
  cd $SUBFOLDER
fi

if [ ! -f .lagoon.yml ]; then
  echo "no .lagoon.yml file found"; exit 1;
fi

echo -e "##############################################\nBEGIN Kubernetes Setup\n##############################################"
sleep 0.5s

if [[ -f "/var/run/secrets/kubernetes.io/serviceaccount/token" ]]; then
  DEPLOYER_TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
else
  if [[ -f "/var/run/secrets/lagoon/deployer/token" ]]; then
    DEPLOYER_TOKEN=$(cat /var/run/secrets/lagoon/deployer/token)
  fi
fi
if [ -z ${DEPLOYER_TOKEN} ]; then
  echo "No deployer token found"; exit 1;
fi

kubectl config set-credentials lagoon/kubernetes.default.svc --token="${DEPLOYER_TOKEN}"
kubectl config set-cluster kubernetes.default.svc --server=https://kubernetes.default.svc --certificate-authority=/run/secrets/kubernetes.io/serviceaccount/ca.crt
kubectl config set-context default/lagoon/kubernetes.default.svc --user=lagoon/kubernetes.default.svc --namespace="${NAMESPACE}" --cluster=kubernetes.default.svc
kubectl config use-context default/lagoon/kubernetes.default.svc

echo -e "\n\n##############################################\nStart Build Process\n##############################################"
.  /kubectl-build-deploy/build-deploy-docker-compose.sh
