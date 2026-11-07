#!/usr/bin/env bash
set -euo pipefail

exec go run ./scripts/mock/import_job_seekers.go -kind recruit "$@"
