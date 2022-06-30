#!/bin/bash
#
# Check the route between tharrr and harrr

set -eu

CURL_DELAY_FILE="curl_in_progress"

echo -n '.'
# Check for the sentinel file. Exit if it's not there.
# If it exists, wait a second, check again, and see if
# it is the SAME sentinel file. If they're different,
# then exit
[[ -f ${CURL_DELAY_FILE} ]] || exit
file_age=$(($(date +%s) - $(date +%s -r "${CURL_DELAY_FILE}")))
if [[ ${file_age} -lt 2 ]]; then
  exit
fi
echo -n -e "\b:"
echo -e "\nSentinel file has existed for more than 1 second"
LOG="route_tr.log"
echo "===" >> ${LOG}
echo "DNS Resolution for api.stripe.com:" >> ${LOG}
/usr/bin/dig +short api.stripe.com >> ${LOG}
THEN=$(date +%s)
echo "Beginning at $(date)" >> ${LOG}
OUTPUT=$(sudo mtr -c 4 -n --report api.stripe.com)
echo "${OUTPUT}" >> ${LOG}
NOW=$(date +%s)
echo "Completed at $(date) after $(( NOW - THEN )) seconds" >> ${LOG}
tail -20 ${LOG}
