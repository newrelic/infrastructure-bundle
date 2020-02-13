# New Relic Infrastructure Bundle

Build tooling to generate and release New Relic Infrastructure containerised **bundle**, including 
on-host agent, service integrations and discovery tools.

## Config

NR will keep latest stable agent and integrations versions at [`versions` file](https://github.com/newrelic/infrastructure-bundle/blob/master/build/versions)

> You could potentially edit the file and set your desired ones at your own risk.

## Build

Run the following command:

   ```bash
   (cd build && make VERSION="<bundle version>")
   ```

## Release

`ci/release.sh` publishes "newrelic/infrastructure-bundle" Docker images triggered *tags* on the *master* branch.

Therefore **[GH Release](https://github.com/newrelic/infrastructure-bundle/release)** is used to trigger TravisCI to deploy into Docker-Hub.

https://hub.docker.com/repository/docker/newrelic/infrastructure-bundle/tags
