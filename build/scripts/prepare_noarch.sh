#!/usr/bin/env sh

set -e
[ -z "$INTEGRATIONS_FILE" ] && INTEGRATIONS_FILE="nri-integrations"

awk -F, '$2 ~ /^bundle$/ && $4 ~ /^noarch$/ {printf "%s_linux_%s_noarch.tar.gz\n",$1,$3;}' ${INTEGRATIONS_FILE}  > ./workspace/nri-noarch
