#!/usr/bin/env bash


set -euo pipefail

TAG=$(echo ${RELEASE_TAG} | sed 's/refs[/]tags[/]//g')

IMAGE="newrelic/infrastructure-bundle"


echo "Scanning Docker image ${IMAGE}:latest ..."
docker run -t -e "SNYK_TOKEN=${SNYK_TOKEN}" -v ${PWD}/workspace:/project -v "/var/run/docker.sock:/var/run/docker.sock" snyk/snyk-cli:docker monitor --docker "${IMAGE}:latest" --severity-threshold=high --org=ohai  --project-name="newrelic/infrastructure-bundle"
