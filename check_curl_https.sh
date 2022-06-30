#!/bin/bash

set -eu

CURL_DELAY_FILE="curl_in_progress"
LOG="curl_https.log"

echo -n '.'
THEN=$(date +%s)
touch "${CURL_DELAY_FILE}"
CURL_OUT=$(curl -s -Iv https://api.stripe.com/healthcheck 2>&1)
NOW=$(date +%s)
if [[ $(( NOW - THEN )) -gt 1 ]]; then
  echo ""
  echo "$(( NOW - THEN )) gt 1"
  echo "===" >> ${LOG}
  echo "${CURL_OUT}" >> ${LOG}
  echo ":: Healthcheck query took $(( NOW - THEN )) seconds to complete (which is > 1 sec)" >> ${LOG}
  tail -40 ${LOG}
fi

