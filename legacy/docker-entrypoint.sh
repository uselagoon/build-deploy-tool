#!/bin/bash
set -e

# try connect to docker-host 10 times before giving up
DOCKER_HOST_COUNTER=1
DOCKER_HOST_TIMEOUT=10
until docker -H ${DOCKER_HOST} info &> /dev/null
do
if [ $DOCKER_HOST_COUNTER -lt $DOCKER_HOST_TIMEOUT ]; then
    let DOCKER_HOST_COUNTER=DOCKER_HOST_COUNTER+1
    echo "${DOCKER_HOST} not available yet, waiting for 5 secs"
    sleep 5
else
    echo "could not connect to ${DOCKER_HOST}"
    exit 1
fi
done


mkdir -p ~/.ssh

cp /var/run/secrets/lagoon/ssh/ssh-privatekey ~/.ssh/key

# Add a new line to the key, as some ssh key formats need a new line
echo "" >> ~/.ssh/key

export SSH_PRIVATE_KEY=$(cat ~/.ssh/key | awk -F'\n' '{if(NR == 1) {printf $0} else {printf "\\n"$0}}')

echo -e "Host * \n    StrictHostKeyChecking no" > ~/.ssh/config
chmod 400 ~/.ssh/*

eval $(ssh-agent)
ssh-add ~/.ssh/key
