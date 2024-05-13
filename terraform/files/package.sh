#!/bin/bash

set -eo pipefail

# Packages flags + env metadata for use by ld-relay.
# Assumes all data and metadata is well-formed.

# first arg is the directory where flags and environment metadata are stored
# second arg is the path + filename of the archive to create. This should end in .tar.gz

if [ -z "$1" ] || [ -z "$2" ]; then
  echo "Error: 2 arguments must be provided."
  exit 1
fi

# Remove existing checksum file and archive
rm -f $1/checksum.md5
rm -f $2

tar -czvf $2 -C $1 .
echo "Generated ld-relay offline mode archive at $(realpath $2)"