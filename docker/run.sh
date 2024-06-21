#!/bin/sh
set -uxe

# poll for s3 object changes in background:
/s3_etag_poll.sh &

# run ld-relay in foreground:
/usr/bin/ldr --from-env