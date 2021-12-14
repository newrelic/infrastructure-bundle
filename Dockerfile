ARG agent_version=latest
ARG base_image=newrelic/infrastructure:${agent_version}

# ephemeral building container
FROM $base_image

# Set by recent docker versions automatically, or by older ones when defining DOCKER_BUILDKIT=1
ARG TARGETOS
ARG TARGETARCH
ARG jre_version

# required for nri-jmx
RUN if [ -n "${jre_version}" ]; then apk add --no-cache openjdk8-jre=${jre_version}; else apk add --no-cache openjdk8-jre; fi

# creating the nri-agent user used only in K8s unprivileged mode
RUN addgroup -g 2000 nri-agent && adduser -D -u 1000 -G nri-agent nri-agent

# integrations
COPY out/${TARGETARCH} /

ENV NRIA_PASSTHROUGH_ENVIRONMENT=ECS_CONTAINER_METADATA_URI,FARGATE
