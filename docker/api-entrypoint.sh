#!/bin/sh
set -eu

mkdir -p /data/uploads /data/exports /data/downloads /data/cisbench
chown -R appuser:appuser /data

exec gosu appuser /usr/local/bin/api
