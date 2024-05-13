#!/bin/sh
set -ue

echo "$(date) Listening for messages on $SQS_QUEUE_URL"
while true; do
  RESULT=$(aws sqs receive-message --queue-url $SQS_QUEUE_URL --wait-time-seconds 20)

  if [ "$RESULT" ]; then
    echo "$(date) Received sqs message. Fetching latest archive from S3"
    RECEIPT_HANDLE=$(echo $RESULT | jq -r '.Messages[0].ReceiptHandle')
    # download archive and delete message from queue
    /pull.sh && aws sqs delete-message --queue-url $SQS_QUEUE_URL --receipt-handle $RECEIPT_HANDLE
  fi
done
