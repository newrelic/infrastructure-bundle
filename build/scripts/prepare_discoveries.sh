#!/usr/bin/env sh

set -e
[ -z "$INTEGRATIONS_FILE" ] && INTEGRATIONS_FILE="nri-integrations"

# example of url
# https://github.com/newrelic/nri-discovery-kubernetes/releases/download/v0.3.0/nri-discovery-kubernetes_0.3.0_Linux_x86_64.tar.gz
awk -F, '$1 ~ /^nri-discovery-/ && $2 ~ /^bundle/ && $4 ~ /^Linux_x86_64/ \
         {printf "https://github.com/newrelic/%s/releases/download/v%s/%s_%s_%s.tar.gz\n",$1,$3,$1,$3,$4;}' \
         ${INTEGRATIONS_FILE}  > ./workspace/nri-discoveries
