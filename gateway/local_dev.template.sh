#!/usr/bin/env bash
set -e

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
pushd ${dir}

export UAA_URL="https://uaa.sys.example.com"
export UAA_USER="v2-nozzle-test"
export UAA_PASS="password"

export ENVELOPES="log,counter,gauge,timer,event"
export LOG_STREAM_URL="https://log-stream.sys.example.com"

export SKIP_SSL_VALIDATION=false

echo "Starting nozzle"
echo ""

go run main.go
