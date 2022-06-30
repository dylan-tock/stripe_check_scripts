#!/bin/bash

set -eu

run() {
  echo "==="
  echo "+ $*"
  date
  "$@"
}
