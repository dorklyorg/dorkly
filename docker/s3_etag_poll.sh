#!/bin/sh
set -ue

ETAG_FILE=/etag.txt
echo "dummy etag" > $ETAG_FILE

echo "$(date) Polling S3 Object's ETag for changes: s3://$S3_BUCKET/flags.tar.gz"
while true; do
  # get the latest ETag from S3

  NEW_ETAG=$(aws s3api head-object --bucket "$S3_BUCKET" --key flags.tar.gz | jq -r '.ETag')
  EXISTING_ETAG=$(cat $ETAG_FILE)
  if [ "$NEW_ETAG" != "$EXISTING_ETAG" ]; then
    echo "$(date) ETag has changed. Fetching latest archive from S3"
    /pull.sh
    echo "$NEW_ETAG" > $ETAG_FILE
  fi
  sleep 5
done
