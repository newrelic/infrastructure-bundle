# `renovate.json`

`renovate.json` is the configuration file for the Renovate bot, which takes care of bumping dependencies periodically.

Reference for the config file can be found here: https://docs.renovatebot.com/configuration-options

This repository makes use of two custom `regexManagers` to automatically update both the base agent image and the integrations when new versions are released.

## `regexManagers`

The first `regexManager` takes care of matching the agent version in the `bundle.yml` file. It does so by using a simple regex `"agentVersion: (?<currentValue>[0-9.]+)"`. Here, `?<currentValue>` creates a group named `currentValue`. This is a special group name defined by Renovate, which identifies the version. The data source (`docker`) and the dependency names (`newrelic/infrastructure`) are hardcoded.

The second `regexManager` is slightly more complex and captures both the name (partially) and the version of the integrations defined in `bundle.yml`. The current version is trivially fetched from the regex just like the one for the agent. The name, however, is composed from the plain repo name in `bundle.yml`, by capturing it into the group `integrationName` and smashing it into the `depNameTemplate`. Finally, the data source is fixed to `github-releases`.

## Other configuration

Two `packageRules` are defined, one to group integration bumps into a single PR, and a second one to pin the `nrjmx` version to the 1.5.x branch.
