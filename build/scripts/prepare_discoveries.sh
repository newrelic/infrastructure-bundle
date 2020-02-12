#!/usr/bin/env sh

set -e

if [ ! -f "$VERSIONS_FILE" ]; then
	echo "VERSIONS_FILE is not set or empty"
	exit 1
fi

if [ ! -d "$WORKSPACE" ]; then
	echo "WORKSPACE is not set or empty"
	exit 1
fi

# example of url
# https://github.com/newrelic/nri-discovery-kubernetes/releases/download/v0.3.0/nri-discovery-kubernetes_0.3.0_Linux_x86_64.tar.gz
awk -F, '$1 ~ /^nri-discovery-/ && $3 ~ /^Linux_x86_64/ \
         {printf "https://github.com/newrelic/%s/releases/download/v%s/%s_%s_%s.tar.gz\n",$1,$2,$1,$2,$3;}' \
         ${VERSIONS_FILE}  > ${WORKSPACE}/nri-discoveries
