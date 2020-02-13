# New Relic Infrastructure Bundle

Build tooling to generate and release New Relic Infrastructure containerised bundle, including 
on-host agent, service integrations and discovery tooling.

## Config

NR will keep latest stable agent and integrations versions at `build/versions`.

> You could potentially edit the file and set your desired ones at your own risk.

## Build

Run the following command:

   ```bash
   make VERSION="<bundle version>"
   ```
