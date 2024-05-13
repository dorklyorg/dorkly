#!/bin/sh
set -uxe
# first time pull of flags archive:
/pull.sh

# listen for sqs messages in background:
/sqs_listen.sh &

# run ld-relay in foreground:
/usr/bin/ldr --from-env