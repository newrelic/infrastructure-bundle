
[![Community Plus header](https://github.com/newrelic/opensource-website/raw/master/src/images/categories/Community_Plus.png)](https://opensource.newrelic.com/oss-category/#community-plus)

# New Relic infrastructure monitoring bundle

Build tooling to generate and release New Relic's containerised **bundle** for infrastructure monitoring, including the on-host agent, service integrations, and discovery tools.

## Requirements

- Docker (with buildx set up if multiarch building is desired)
- Go (for the downloader)
- *nix compatible shell (verified to work on Linux and OSX)

## Installation and usage

For install instructions, see [Docker container for infrastructure monitoring](https://docs.newrelic.com/docs/infrastructure/install-infrastructure-agent/linux-installation/docker-container-infrastructure-monitoring).

## Configuration

New Relic will keep latest stable agent and integrations versions in the [`bundle.yml` file](https://github.com/newrelic/infrastructure-bundle/blob/master/bundle.yml).

> You can edit the file and set your desired versions, at your own risk.

## Building

Building multiarch images requires a working setup of [docker buildx](https://docs.docker.com/buildx/working-with-buildx/).
A working installation of Go is also needed for running the downloader program.

```bash
DOCKER_PLATFORMS=linux/amd64 ./run-ci-locally.sh
```

### Without `docker buildx`

A single-arch image can also be built without `buildx`. However, setting `DOCKER_BUILDKIT=1` might be required for older versions of docker, otherwise the `TARGETOS` and `TARGETARCH` variables won't be populated and docker will fail to copy the integrations from the host.

```bash
# Run downloader script
go run downloader.go

# Build image
DOCKER_BUILDKIT=1 docker build . -t newrelic/infrastructure-bundle:dev
```

## Release

Follow this [guide](https://github.com/newrelic/coreint-automation/blob/main/docs/release_runbook.md) to create a release.

### Released Images
CI workflow pushes the multiarch image to [dockerhub](https://hub.docker.com/repository/docker/newrelic/infrastructure-bundle/tags) by running `docker buildx` with `--push` whenever a release or prerelease is published in Github. The tag is generated from the release tag (after stripping the leading `v`). Additionally, prereleases will have an `-rc` suffix automatically appended to the tag.

Locally, this can be also be done with the `./run-ci-locally.sh` script:

```bash
DOCKER_IMAGE_TAG=0.0.1-rc ./run-ci-locally.sh release
```


## Bumping versions

**Note** Renovate should take care of bumping versions, and auto-merge them when all checks passes. If so automatically releases will pick that versions bumps without needed to add any changelog.

Versions, urls, and architectures of the bundled integrations are defined in `bundle.yml`.

The version of the [base agent image](https://hub.docker.com/repository/docker/newrelic/infrastructure/tags) is also defined in `bundle.yml`, and is collected by the `docker-build.sh` wrapper script.

## Support

Should you need assistance with New Relic products, you are in good hands with several support diagnostic tools and support channels.

If the issue has been confirmed as a bug or is a feature request, file a GitHub issue.

**Support Channels**

* [New Relic Documentation](https://docs.newrelic.com): Comprehensive guidance for using our platform
* [New Relic Community](https://discuss.newrelic.com/c/support-products-agents/new-relic-infrastructure): The best place to engage in troubleshooting questions
* [New Relic Developer](https://developer.newrelic.com/): Resources for building a custom observability applications
* [New Relic University](https://learn.newrelic.com/): A range of online training for New Relic users of every level
* [New Relic Technical Support](https://support.newrelic.com/) 24/7/365 ticketed support. Read more about our [Technical Support Offerings](https://docs.newrelic.com/docs/licenses/license-information/general-usage-licenses/support-plan).

## Privacy

At New Relic we take your privacy and the security of your information seriously, and are committed to protecting your information. We must emphasize the importance of not sharing personal data in public forums, and ask all users to scrub logs and diagnostic information for sensitive information, whether personal, proprietary, or otherwise.

We define “Personal Data” as any information relating to an identified or identifiable individual, including, for example, your name, phone number, post code or zip code, Device ID, IP address, and email address.

For more information, review [New Relic’s General Data Privacy Notice](https://newrelic.com/termsandconditions/privacy).

## Contribute

We encourage your contributions to improve this project! Keep in mind that when you submit your pull request, you'll need to sign the CLA via the click-through using CLA-Assistant. You only have to sign the CLA one time per project.

If you have any questions, or to execute our corporate CLA (which is required if your contribution is on behalf of a company), drop us an email at opensource@newrelic.com.

**A note about vulnerabilities**

As noted in our [security policy](../../security/policy), New Relic is committed to the privacy and security of our customers and their data. We believe that providing coordinated disclosure by security researchers and engaging with the security community are important means to achieve our security goals.

If you believe you have found a security vulnerability in this project or any of New Relic's products or websites, we welcome and greatly appreciate you reporting it to New Relic through [HackerOne](https://hackerone.com/newrelic).

If you would like to contribute to this project, review [these guidelines](./CONTRIBUTING.md).

To all contributors, we thank you!  Without your contribution, this project would not be what it is today.

## License

infrastructure-bundle is licensed under the [Apache 2.0](http://apache.org/licenses/LICENSE-2.0.txt) License.
