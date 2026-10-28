ARG agent_version=latest
ARG base_image_name=newrelic/infrastructure
ARG base_image=${base_image_name}:${agent_version}

# ephemeral building container
FROM $base_image

# Set by recent docker versions automatically, or by older ones when defining DOCKER_BUILDKIT=1
ARG TARGETOS
ARG TARGETARCH
ARG jre_version
ARG out_dir

# Refresh inherited base-image OS packages so transient Alpine CVEs (e.g. curl,
# libexpat) are patched even when the published base image predates the fix.
# The base newrelic/infrastructure image runs the same upgrade at build time.
RUN apk --no-cache upgrade

# required for nri-jmx
# Alpine only ships openjdk8-jre for 32-bit arm (armv7); newer JREs exist for
# amd64/arm64 only. Newer JREs also avoid the krb5-libs dependency of openjdk8.
RUN case "${TARGETARCH}" in \
      arm) jre_package=openjdk8-jre ;; \
      *)   jre_package=openjdk17-jre-headless ;; \
    esac; \
    if [ -n "${jre_version}" ]; then apk add --no-cache ${jre_package}=${jre_version}; else apk add --no-cache ${jre_package}; fi

# integrations
COPY ${out_dir}/${TARGETARCH} /
