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

awk -F, '$3 ~ /^noarch$/ {printf "%s_linux_%s_noarch.tar.gz\n",$1,$2;}' ${VERSIONS_FILE}  > ${WORKSPACE}/nri-noarch
