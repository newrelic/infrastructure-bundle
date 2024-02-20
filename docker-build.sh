#!/usr/bin/env sh

# Wrapper script for docker buildx build setting default values

DOCKER_PLATFORMS=${DOCKER_PLATFORMS:-linux/amd64,linux/arm64,linux/arm}
JRE_VERSION=${JRE_VERSION:-} # Blank will pull default version for alpine image
DOCKER_IMAGE=${DOCKER_IMAGE:-newrelic/infrastructure-bundle}
DOCKER_IMAGE_TAG=${DOCKER_IMAGE_TAG:-dev} # Overwritten by CI from the release tag

# Get default AGENT_VERSION from downloader.go
if [ -z "$AGENT_VERSION" ]; then
    AGENT_VERSION=$(go run ./downloader.go -agent-version)
    if [ -z "$AGENT_VERSION" ]; then
        echo "Could not get agent version from downloader.go" >&2
        exit 1
    fi
fi

echo Building the image leveraging agent_version=$AGENT_VERSION

docker buildx build \
  --platform="${DOCKER_PLATFORMS}" \
  --build-arg agent_version="$AGENT_VERSION" \
  --build-arg jre_version="$JRE_VERSION" \
  -t "${DOCKER_IMAGE}:${DOCKER_IMAGE_TAG}" \
  "$@"
