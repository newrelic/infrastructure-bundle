#!/bin/bash

set -euo pipefail

echo "Building version $RELEASE_TAG ..."

(make -C ./build/ VERSION="$RELEASE_TAG" )

echo "Docker logging in ..."

docker login -u $DOCKER_USERNAME -p $DOCKER_PASSWORD

echo "Releasing ..."

IMAGE="newrelic/infrastructure-bundle"
docker tag ${IMAGE}:${RELEASE_TAG} ${IMAGE}:latest
docker push ${IMAGE}:${RELEASE_TAG}
docker push ${IMAGE}:latest
