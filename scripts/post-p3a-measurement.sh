#!/bin/bash

endpoint="https://nitro.nymity.ch:8080/report"

read -r -d '' prefix <<- EOM
{
  "channel":"developer",
  "country_code":"
EOM

read -r -d '' suffix<<- EOM
  ",
  "metric_name":"Brave.Welcome.InteractionStatus",
  "metric_value":3,
  "platform":"linux-bc",
  "refcode":"none",
  "version":"1.36.68",
  "woi":4,
  "wos":4,
  "yoi":2022,
  "yos":2022
}
EOM

function send_request() {
    json_blob="$1"
    curl --insecure \
        -X POST \
        -H "Content-Type: application/json" \
        -i \
        -d "$json_blob" \
        "$endpoint"
}

# Send enough requests to be above the crowd ID threshold.
for ((i=0; i<10; i++))
do
    send_request "${prefix}US${suffix}\n"
done

# Send a few requests that are below the crowd ID threshold, and will be
# discarded by the shuffler.
for ((i=0; i<5; i++))
do
    send_request "${prefix}CA${suffix}\n"
done
