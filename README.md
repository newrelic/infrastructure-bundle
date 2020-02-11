# New Relic Infrastructure Bundle

Build tooling to generate and release New Relic Infrastructure containerised bundle, including 
on-host agent, service integrations and discovery tooling.

## Build

1. Get your desired containerised Infrastructure agent from the tags avaiable at
[`newrelic/infrastructure`](https://hub.docker.com/r/newrelic/infrastructure/tags)

1. NR will keep latest stable integrations versions at `build/nri-integrations`.
   > You could potentially edit the file and set your desired ones at your own risk.

1. Run the following command:

   ```bash
   make VERSION=<bundle version> AGENT_VER=<version of agent> 
   ```
