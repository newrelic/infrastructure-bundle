#!/usr/bin/env sh

set -euo pipefail

ws="/tmp/workspace"
mkdir -p ${ws}
cd ${ws}

# File contains file names
while IFS= read -r file
do
    url="https://download.newrelic.com/infrastructure_agent/binaries/linux/amd64/${file}"
    echo "Getting integration ${url}"
    curl --location --fail --silent "${url}" --output "${ws}/${file}"
    tar -xzf "${ws}/${file}"
done < "/etc/nri-integrations"

# File contains file names
while IFS= read -r file
do
    url="https://download.newrelic.com/infrastructure_agent/binaries/linux/noarch/${file}"
    echo "Getting noarch ${url}"
    curl --location --fail --silent "${url}" --output "${ws}/${file}"
    mkdir -p ${ws}/aux
    tar -xzf "${ws}/${file}" -C ${ws}/aux
    cp -r ${ws}/aux/* /
    rm -rf ${ws}/aux
done < "/etc/nri-noarch"

# File contains URLs
while IFS= read -r url
do
    echo "Getting discovery ${url}"
    curl --location --fail --silent "${url}" --output "${ws}/dicovery.tar.gz"
    mkdir -p ${ws}/aux
    tar -xzf "${ws}/dicovery.tar.gz" -C ${ws}/aux
    cp -r ${ws}/aux/* /var/db/newrelic-infra/
    rm -rf ${ws}/aux
done < "/etc/nri-discoveries"

# cleanup tars, config sample files, windows definition files
rm -rf ${ws}/*.tar.gz
rm -rf ${ws}/etc/newrelic-infra/integrations.d/*.sample
rm -rf ${ws}/var/db/newrelic-infra/newrelic-integrations/*-win-*.yml

# copy to proper location
# this could overwrite sys files in case of upstream bug, but just on the intermediate builder
cp -r  ${ws}/* /
