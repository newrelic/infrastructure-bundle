ARG agent_version
ARG base_image=newrelic/infrastructure:${agent_version}

# ephemeral building container
FROM $base_image AS builder

RUN apk add --no-cache curl

COPY workspace/nri-integrations /etc/
COPY workspace/nri-discoveries /etc/
COPY workspace/nri-noarch /etc/
COPY scripts/install.sh /tmp/

RUN chmod +x /tmp/install.sh \
    && ./tmp/install.sh

# infrastructure-bundle image
FROM $base_image AS bundle
ARG jre_version

# required for nri-jmx
RUN apk add --no-cache openjdk8-jre=${jre_version}

# integrations
COPY --from=builder /var/db/newrelic-infra /var/db/newrelic-infra/
COPY --from=builder /etc/nri-* /etc/
COPY --from=builder /usr/bin/nrjmx* /usr/bin/

# LABELS_PLACEHOLDER
