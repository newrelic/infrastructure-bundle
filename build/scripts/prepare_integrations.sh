#!/usr/bin/env sh

set -e
[ -z "$INTEGRATIONS_FILE" ] && INTEGRATIONS_FILE="nri-integrations"

awk -F, '$2 ~ /^bundle$/ && $4 ~ /^$|^amd64$/ {printf "%s_linux_%s_amd64.tar.gz\n",$1,$3;}' ${INTEGRATIONS_FILE}  > ./workspace/nri-integrations
