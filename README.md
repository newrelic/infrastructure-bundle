# New Relic Infrastructure Bundle

Build tooling to generate and release New Relic Infrastructure containerised **bundle**, including 
on-host agent, service integrations and discovery tools.

Contained versions are available at https://github.com/newrelic/infrastructure-bundle/blob/master/build/versions

This repo releases "newrelic/infrastructure-bundle" Docker images triggered *tags* on the *master* branch. Therefore "GH Release" is aimed to be used to publish into Docker-Hub.

https://hub.docker.com/repository/docker/newrelic/infrastructure-bundle/tags

## Config

NR will keep latest stable agent and integrations versions at `build/versions`.

> You could potentially edit the file and set your desired ones at your own risk.

## Build

Run the following command:

   ```bash
   make VERSION="<bundle version>"
   ```
