#!/usr/bin/env bash

set -e

#  Environment variables:
#    - Required:
#      - AGENT_VERSION: Infra agent version that's being added.
#      - IMAGE_TAG: tag for the Docker image.
#      - WORKSPACE: Local workspace folder for the builder to fetch data from.

# Ensure AGENT_VERSION is set & non-empty
if [ -z "$AGENT_VERSION" ]; then
	echo "AGENT_VERSION is not set or empty"
	exit 1
fi

# Ensure IMAGE_TAG is set & non-empty
if [ -z "$IMAGE_TAG" ]; then
	echo "IMAGE_TAG is not set or empty"
	exit 1
fi

# Ensure WORKSPACE is set & non-empty
if [ -z "$WORKSPACE" ]; then
	echo "WORKSPACE is not set or empty"
	exit 1
fi

docker build \
	--no-cache \
	-t $IMAGE_TAG \
	--build-arg agent_version=$AGENT_VERSION \
	-f Dockerfile \
	.
