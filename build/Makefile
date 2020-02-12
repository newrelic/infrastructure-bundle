.DEFAULT_GOAL := default

# meant to be replaced by CI
VERSION           ?= dev
AGENT_VER         ?= latest
# usually fixed
TAG               ?= newrelic/infrastructure-bundle:${VERSION}
WORKSPACE_DIR     ?= ./workspace
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
prepare-integrations/nri-integrations : export WORKSPACE = ${WORKSPACE_DIR}
prepare-integrations/nri-integrations :
	@(./scripts/prepare_integrations.sh)

prepare-integrations/nri-noarch : workspace
prepare-integrations/nri-noarch : export WORKSPACE = ${WORKSPACE_DIR}
prepare-integrations/nri-noarch :
	@(./scripts/prepare_noarch.sh)

prepare-integrations/nri-discoveries : workspace
prepare-integrations/nri-discoveries : export WORKSPACE = ${WORKSPACE_DIR}
prepare-integrations/nri-discoveries :
	@(./scripts/prepare_discoveries.sh)

.PHONY : prepare-labels
prepare-labels :
	@(awk -F, '{if (NF==2) {printf "com.newrelic.%s.version=%s ",$$1,$$2}}' ${INTEGRATIONS_FILE} | xargs -I{} sed "s/\# LABELS_PLACEHOLDER/LABEL {}/g" ./Dockerfile > ${WORKSPACE_DIR}/Dockerfile)


# Helpers

.PHONY : lint
	@(docker run --rm -i hadolint/hadolint < Dockerfile)

