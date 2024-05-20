#!/bin/sh
set -ux

# Silently fail if the object does not exist. This is useful for the first run before the archive is uploaded to S3.
aws s3 cp "$S3_URL" /dorkly/flags.tar.gz || true