#!/usr/bin/env bash
set -euo pipefail

dirs=$(go list -f '{{.Dir}}/...' -m | xargs)
./bin/go-licenses report $dirs --template ./tools/licenses.tpl > LICENSE-3rdparty.csv 2> errors
