#!/bin/sh
set -eu

: "${PHONE:?set PHONE to an E.164 phone number}"
: "${CAPTCHA_TOKEN:?set CAPTCHA_TOKEN}"
: "${WIDGET_RECORD_ID:?set WIDGET_RECORD_ID}"

curl --fail-with-body -X POST http://localhost:8080/login/code \
  -H 'Content-Type: application/json' \
  -d "{\"phone\":\"${PHONE}\",\"captcha_token\":\"${CAPTCHA_TOKEN}\",\"widget_record_id\":\"${WIDGET_RECORD_ID}\"}"

printf '\nEnter the SMS code: '
read -r CODE
curl --fail-with-body -X POST http://localhost:8080/login/verify \
  -H 'Content-Type: application/json' \
  -d "{\"phone\":\"${PHONE}\",\"code\":\"${CODE}\",\"order_id\":\"ord_1042\"}"
