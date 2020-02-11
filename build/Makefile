.DEFAULT_GOAL := default

# meant to be replaced by CI
VERSION       ?= dev
AGENT_VER     ?= latest
# usually fixed
TAG           ?= newrelic/infrastructure-bundle:${VERSION}
WORKSPACE_DIR ?= ./workspace
LABELS_FILE   ?= ${WORKSPACE_DIR}/bundle_labels
INTEGRATIONS_FILE ?= nri-integrations

.PHONY : clean
default : clean build

.PHONY : clean
clean :
	@(rm -rf workspace)

.PHONY : build
build : workspace
build : prepare-integrations
build : prepare-labels
build : export AGENT_VERSION = ${AGENT_VER}
build : export IMAGE_TAG = ${TAG}
build : export WORKSPACE = ${WORKSPACE_DIR}
build :
	@(./build.sh)

.PHONY : workspace
workspace :
	@(mkdir -p ${WORKSPACE_DIR})

.PHONY : prepare-integrations
prepare-integrations : prepare-integrations/nri-integrations
prepare-integrations : prepare-integrations/nri-discoveries
prepare-integrations : prepare-integrations/nri-noarch

prepare-integrations/nri-integrations : workspace
	@(./scripts/prepare_integrations.sh)

prepare-integrations/nri-noarch : workspace
	@(./scripts/prepare_noarch.sh)

prepare-integrations/nri-discoveries : workspace
	@(./scripts/prepare_discoveries.sh)

.PHONY : prepare-labels
prepare-labels :
	awk -F, '$$2 ~ /^bundle$$/ {printf "com.newrelic.%s.version=%s ",$$1,$$3}' ${INTEGRATIONS_FILE} > ${LABELS_FILE}
#	(sed -i 's/\# LABELS_PLACEHOLDER/LABEL $(shell cat ${LABELS_FILE})/g' Dockerfile)


# Helpers

.PHONY : lint
	@(docker run --rm -i hadolint/hadolint < Dockerfile)

