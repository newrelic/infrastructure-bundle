#!/usr/bin/env bash

set -euo pipefail

TAG=$(echo ${RELEASE_TAG} | sed 's/refs[/]tags[/]//g')

echo "Building version $TAG ..."

(make -C ./build/ VERSION="$TAG" )

echo "Docker logging in ..."

docker login -u $DOCKER_USERNAME -p $DOCKER_PASSWORD

echo "Releasing ..."

IMAGE="newrelic/infrastructure-bundle"
docker tag ${IMAGE}:${TAG} ${IMAGE}:latest
docker push ${IMAGE}:${TAG}
docker push ${IMAGE}:latest
