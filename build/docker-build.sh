#!/usr/bin/env sh

# Wrapper script for docker buildx build setting default values

DOCKER_PLATFORMS=${DOCKER_PLATFORMS:-linux/amd64,linux/arm64,linux/arm}
AGENT_VERSION=${AGENT_VERSION:-latest}
JRE_VERSION=${JRE_VERSION:-} # Blank will pull default version for alpine image
DOCKER_IMAGE=${DOCKER_IMAGE:-newrelic/infrastructure-bundle}
DOCKER_IMAGE_TAG=${DOCKER_IMAGE_TAG:-dev} # Overwritten by CI from the release tag

docker buildx build \
  --platform="${DOCKER_PLATFORMS}" \
  --build-arg agent_version="$AGENT_VERSION" \
  --build-arg jre_version="$JRE_VERSION" \
  -t "${DOCKER_IMAGE}:${DOCKER_IMAGE_TAG}" \
  "$@"
