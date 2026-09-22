#!/bin/sh
set -eu
script_dir=$(CDPATH= cd -P "$(dirname "$0")" && pwd)
exec "${PYTHON:-python3}" "$script_dir/api.py" --url "${ZENITH_URL:-http://127.0.0.1:8080}" status "$@"
