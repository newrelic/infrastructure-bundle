#!/usr/bin/env bash

set -e

# This scripts runs something analogous to the CI locally, for reproducibility.
# If $1 is set to release, it will run the release job with the docker account currently logged in

echo "Sourcing env from ./buildsettings.env"
source buildsettings.env

DOCKER_TAG=${DOCKER_TAG:-dev}

echo "Downloading integrations..."
export GO111MODULE=auto
go get gopkg.in/yaml.v3
go run downloader.go

echo "Building image..."
docker buildx build . --platform="${DOCKER_PLATFORMS}"
docker buildx build . --platform=linux/amd64 --load -t "${DOCKER_IMAGE}:${DOCKER_TAG}"

if [[ -n "$SNYK_TOKEN" ]]; then
    echo "Scanning Docker image ${DOCKER_IMAGE}:${DOCKER_TAG}..."
    docker run -t -e "SNYK_TOKEN=${SNYK_TOKEN}" -v ${PWD}/workspace:/project -v "/var/run/docker.sock:/var/run/docker.sock" snyk/snyk-cli:docker monitor --docker "${DOCKER_IMAGE}:${DOCKER_TAG}" --severity-threshold=high --org=ohai  --project-name="${DOCKER_IMAGE}"
else
    echo "SNYK_TOKEN not defined, skipping snyk check"
fi

if [[ $1 == "release" ]]; then
    if [[ ${DOCKER_TAG} = "dev" ]]; then
        echo "Refusing to push image with default tag '${DOCKER_TAG}'. Please override DOCKER_TAG env."
        exit 1
    fi

    echo
    echo "Will now build and push ${DOCKER_IMAGE}:$gittag as the user currently logged in in docker."
    echo "If this is not what you want, press ^C within 5 seconds..."
    echo
    sleep 5
    docker buildx build --bush . --platform="${DOCKER_PLATFORMS}"
fi
