#!/bin/bash

set -euo pipefail

echo "Building version $TRAVIS_TAG ..."

(cd ../build; make VERSION="$TRAVIS_TAG" )

echo "Docker logging in ..."

docker login -u $DOCKER_USERNAME -p $DOCKER_PASSWORD

echo "Releasing ..."

IMAGE="newrelic/infrastructure-bundle"
docker tag ${IMAGE}:${TRAVIS_TAG} ${IMAGE}:latest
docker push ${IMAGE}:${TRAVIS_TAG}
docker push ${IMAGE}:latest
