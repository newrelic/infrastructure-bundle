#!/usr/bin/env bash

# Wrapper script for Windows container builds
# NOTE: This must be run on a Windows host with Docker configured for Windows containers

set -e

WINDOWS_VERSION=${WINDOWS_VERSION:-ltsc2022}
DOCKER_IMAGE=${DOCKER_IMAGE:-newrelic/infrastructure-bundle}
DOCKER_IMAGE_TAG=${DOCKER_IMAGE_TAG:-dev}
OUTDIR=${OUTDIR:-out-windows}
BASE_IMAGE_NAME=${BASE_IMAGE_NAME:-newrelic/infrastructure}

echo "Building Windows image for Windows Server $WINDOWS_VERSION"
echo "base_image_name: $BASE_IMAGE_NAME"

# Get default AGENT_VERSION from downloader.go
if [ -z "$AGENT_VERSION" ]; then
    AGENT_VERSION=$(go run ./downloader.go -bundle=bundle-windows.yml -agent-version)
    if [ -z "$AGENT_VERSION" ]; then
        echo "Could not get agent version from downloader.go" >&2
        exit 1
    fi
fi

echo "Building the image leveraging agent_version=$AGENT_VERSION for Windows Server $WINDOWS_VERSION"

# Construct the base image tag to include Windows version
BASE_IMAGE_TAG="${AGENT_VERSION}-${WINDOWS_VERSION}"
FULL_TAG="${DOCKER_IMAGE}:${DOCKER_IMAGE_TAG}-${WINDOWS_VERSION}"

# Check if --push flag is passed
PUSH_IMAGE=false
for arg in "$@"; do
  if [ "$arg" = "--push" ]; then
    PUSH_IMAGE=true
  fi
done

# Remove --push from arguments as it's not valid for docker build
FILTERED_ARGS=()
for arg in "$@"; do
  if [ "$arg" != "--push" ]; then
    FILTERED_ARGS+=("$arg")
  fi
done

docker build \
  --build-arg agent_version="$BASE_IMAGE_TAG" \
  --build-arg base_image_name="$BASE_IMAGE_NAME" \
  --build-arg out_dir="$OUTDIR" \
  -t "${FULL_TAG}" \
  -f Dockerfile.windows \
  . \
  "${FILTERED_ARGS[@]}"

echo "Successfully built: ${FULL_TAG}"

# Push if --push was requested
if [ "$PUSH_IMAGE" = true ]; then
  echo "Pushing ${FULL_TAG}..."
  docker push "${FULL_TAG}"
  echo "Successfully pushed: ${FULL_TAG}"
fi
