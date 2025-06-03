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

# required for nri-jmx
RUN if [ -n "${jre_version}" ]; then apk add --no-cache openjdk8-jre=${jre_version}; else apk add --no-cache openjdk8-jre; fi

# integrations
COPY ${out_dir}/${TARGETARCH} /
