#!/bin/sh
set -uxe
aws s3 cp "$S3_URL" /dorkly/flags.tar.gz
