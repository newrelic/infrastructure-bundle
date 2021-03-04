#!/usr/bin/env bash

set -e

# This scripts runs something analogous to the CI locally, for reproducibility.
# If $1 is set to release, it will run the release job with the docker account currently logged in

echo "Downloading integrations..."
export GO111MODULE=auto
go get gopkg.in/yaml.v3
go run downloader.go

echo "Building image..."
# Test build, but not load, for all archs
./docker-build.sh .
# Build and load for amd64 only (for snyk)
DOCKER_PLATFORMS=linux/amd64 ./docker-build.sh . --load

if [[ -n "$SNYK_TOKEN" ]]; then
    echo "Scanning Docker image ${DOCKER_IMAGE}:${DOCKER_TAG}..."
    docker run -t -e "SNYK_TOKEN=${SNYK_TOKEN}" -v ${PWD}/workspace:/project -v "/var/run/docker.sock:/var/run/docker.sock" snyk/snyk-cli:docker monitor --docker "${DOCKER_IMAGE}:${DOCKER_TAG}" --severity-threshold=high --org=ohai  --project-name="${DOCKER_IMAGE}"
else
    echo "SNYK_TOKEN not defined, skipping snyk check"
fi

if [[ $1 == "release" ]]; then
    if [[ -z ${DOCKER_IMAGE_TAG} ]]; then
        echo "Refusing to push image with default tag. Please set the DOCKER_IMAGE_TAG env var."
        exit 1
    fi

    echo
    echo "Will now build and push ${DOCKER_IMAGE}:${DOCKER_IMAGE_TAG} as the user currently logged in in the docker CLI."
    echo "If this is not what you want, press ^C within 5 seconds..."
    echo
    sleep 5
    ./docker-build.sh . --push
fi
