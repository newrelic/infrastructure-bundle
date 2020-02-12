#!/usr/bin/env sh

set -e
[ -z "$INTEGRATIONS_FILE" ] && INTEGRATIONS_FILE="nri-integrations"

if [ -z "$WORKSPACE" ]; then
	echo "WORKSPACE is not set or empty"
	exit 1
fi

awk -F, '$3 ~ /^$/ {printf "%s_linux_%s_amd64.tar.gz\n",$1,$2;}' ${INTEGRATIONS_FILE}  > ${WORKSPACE}/nri-integrations
