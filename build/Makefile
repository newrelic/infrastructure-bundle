.DEFAULT_GOAL := default

# meant to be replaced by CI
VERSION           ?= dev
# usually fixed
TAG               ?= newrelic/infrastructure-bundle:${VERSION}
WORKSPACE_DIR     ?= ./workspace
VERSIONS_FILE_    := versions

.PHONY : default
default : clean build

.PHONY : clean
clean :
	@(rm -rf ${WORKSPACE_DIR})

.PHONY : build
build : prepare-workspace
build : prepare-integrations
build : prepare-labels
build : export IMAGE_TAG     = ${TAG}
build : export WORKSPACE     = ${WORKSPACE_DIR}
build : export VERSIONS_FILE = ${VERSIONS_FILE_}
build :
	@(./scripts/build.sh)

.PHONY : prepare-workspace
prepare-workspace :
	@(mkdir -p ${WORKSPACE_DIR})

.PHONY : prepare-integrations
prepare-integrations : prepare-integrations/nri-integrations
prepare-integrations : prepare-integrations/nri-discoveries
prepare-integrations : prepare-integrations/nri-noarch

prepare-integrations/nri-integrations : workspace
prepare-integrations/nri-integrations : export WORKSPACE = ${WORKSPACE_DIR}
prepare-integrations/nri-integrations : export VERSIONS_FILE = ${VERSIONS_FILE_}
prepare-integrations/nri-integrations :
	@(./scripts/prepare_integrations.sh)

prepare-integrations/nri-noarch : workspace
prepare-integrations/nri-noarch : export WORKSPACE = ${WORKSPACE_DIR}
prepare-integrations/nri-noarch : export VERSIONS_FILE = ${VERSIONS_FILE_}
prepare-integrations/nri-noarch :
	@(./scripts/prepare_noarch.sh)

prepare-integrations/nri-discoveries : workspace
prepare-integrations/nri-discoveries : export WORKSPACE = ${WORKSPACE_DIR}
prepare-integrations/nri-discoveries : export VERSIONS_FILE = ${VERSIONS_FILE_}
prepare-integrations/nri-discoveries :
	@(./scripts/prepare_discoveries.sh)

.PHONY : prepare-labels
prepare-labels :
	@(awk -F, '{if (NF==2 && $$1 ~ /^nri-.*/) {printf "com.newrelic.%s.version=%s ",$$1,$$2}}' ${VERSIONS_FILE_} | xargs -I{} sed "s/\# LABELS_PLACEHOLDER/LABEL {}/g" ./Dockerfile > ${WORKSPACE_DIR}/Dockerfile)


# Helpers

.PHONY : lint
lint :
	@(docker run --rm -i hadolint/hadolint < Dockerfile)

