#!/bin/bash
#
# Items we only need run one time

set -eu

source check_common.sh

check_os() {
    uname="$(uname)"
    case "$uname" in
        Linux|Darwin)
            ;;
        *)
            echo "WARNING: not tested on $uname"
            return 1
            ;;
    esac
  echo "Running on os '${uname}' at $(date)"
}

check_ip() {
  run curl -4 --write-out "\n" ifconfig.co/json  2> /dev/null
}

dig_short() {
    output="$(run dig +short "$@")"
    if [ -z "$output" ]; then
        echo >&2 "Error: command returned no output: dig +short $*"
        return 1
    fi
    echo "$output"
}

gethostbyname() {
    run python3 -c "from socket import gethostbyname; print(gethostbyname('$1'))"
}

check_os
gethostbyname api.stripe.com
dig_short api.stripe.com
check_ip

