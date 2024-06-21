# tag should be kept in sync with internal/dorkly/integration_test.go:22
FROM launchdarkly/ld-relay:8.4.2
LABEL authors="drichelson"

USER root
RUN apk add aws-cli jq --no-cache

# This env var is consumed by ld-relay and its presence is required to start the relay in offline mode
ENV FILE_DATA_SOURCE=/dorkly/flags.tar.gz

# Sets ld-relay log level
ENV LOG_LEVEL=info

COPY flags.tar.gz /dorkly/flags.tar.gz
COPY pull.sh /pull.sh
COPY run.sh /run.sh
COPY s3_etag_poll.sh /s3_etag_poll.sh
RUN chmod +x /pull.sh /run.sh /s3_etag_poll.sh

ENTRYPOINT ["/run.sh"]