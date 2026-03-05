#!/bin/sh
set -eu

mkdir -p /data/uploads /data/exports
chown -R appuser:appuser /data

exec su-exec appuser /usr/local/bin/api
