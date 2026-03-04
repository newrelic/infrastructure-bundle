#!/usr/bin/env sh

# Wrapper script for building Windows docker images.
# Windows containers do not support cross-compilation via buildx,
# so this uses plain `docker build` and must run on a Windows host
# with Docker switched to Windows containers mode.

DOCKER_IMAGE=${DOCKER_IMAGE:-newrelic/infrastructure-bundle}
DOCKER_IMAGE_TAG=${DOCKER_IMAGE_TAG:-dev}
WINDOWS_VERSION=${WINDOWS_VERSION:-ltsc2019}
BASE_IMAGE_NAME=${BASE_IMAGE_NAME:-newrelic/infrastructure-windows}

echo "base_image_name $BASE_IMAGE_NAME\n"

# Get default AGENT_VERSION from downloader.go
if [ -z "$AGENT_VERSION" ]; then
    AGENT_VERSION=$(go run ./downloader.go -agent-version)
    if [ -z "$AGENT_VERSION" ]; then
        echo "Could not get agent version from downloader.go" >&2
        exit 1
    fi
fi

echo "Building Windows image for ${WINDOWS_VERSION} with agent_version=${AGENT_VERSION}"

# Parse --push from args since plain `docker build` does not support it (unlike buildx)
PUSH=false
for arg in "$@"; do
    if [ "$arg" = "--push" ]; then
        PUSH=true
    fi
done

FULL_TAG="${DOCKER_IMAGE}:${DOCKER_IMAGE_TAG}-servercore-${WINDOWS_VERSION}"

docker build \
  --build-arg agent_version="${AGENT_VERSION}" \
  --build-arg windows_version="${WINDOWS_VERSION}" \
  --build-arg base_image_name="${BASE_IMAGE_NAME}" \
  -t "${FULL_TAG}" \
  -f Dockerfile.windows .

if [ "$PUSH" = "true" ]; then
    docker push "${FULL_TAG}"
fi
