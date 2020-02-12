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


awk -F, '$1 ~ /^nri-$/ {printf "%s_linux_%s_amd64.tar.gz\n",$1,$2;}' ${VERSIONS_FILE}  > ${WORKSPACE}/nri-integrations
