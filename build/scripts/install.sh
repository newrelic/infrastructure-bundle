#!/usr/bin/env sh

set -euo pipefail

input="/etc/nri-integrations"
while IFS= read -r binary
do
    int_url="https://download.newrelic.com/infrastructure_agent/binaries/linux/amd64/${binary}"
    echo "Getting ${int_url}"
    curl --location --fail --silent "${int_url}" --output "/tmp/${binary}"
    tar -xzf "/tmp/${binary}"
done < $input
# cleanup
rm -rf /etc/newrelic-infra/integrations.d/*.sample
# windows definition files
rm -rf /var/db/newrelic-infra/newrelic-integrations/*-win-*.yml
rm -rf /tmp/**

input="/etc/nri-noarch"
while IFS= read -r binary
do
    int_url="https://download.newrelic.com/infrastructure_agent/binaries/linux/noarch/${binary}"
    echo "Getting ${int_url}"
    curl --location --fail --silent "${int_url}" --output "/tmp/${binary}"
    tar -xzf "/tmp/${binary}"
done < $input
# cleanup
rm -rf /etc/newrelic-infra/integrations.d/*.sample
rm -rf /tmp/**

input="/etc/nri-discoveries"
while IFS= read -r discovery
do
    echo "Getting ${discovery}"
    curl --location --fail --silent "${discovery}" --output "/tmp/dicovery.tar.gz"
    mkdir -p /tmp/binary
    tar -C /tmp/binary -xzf "/tmp/dicovery.tar.gz"
    cp /tmp/binary/nri-* /var/db/newrelic-infra/
    rm -rf /tmp/binary
done < $input
