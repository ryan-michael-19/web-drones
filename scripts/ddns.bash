#!/usr/bin/env bash

set -euf -o pipefail

# TODO: Move ./api_key to ./secret
API_KEY=$(cat ./api_key)
CURRENT_IP=$(cat ./current_ip)
ZONE_ID=$(cat ./zone_id)

EXTERNAL_IP=$(curl -s "https://api.ipify.org")
if [ -f ./current_ip ] ; then
    CURRENT_IP=$(cat current_ip)
else
    echo "./current_ip does not exist, creating it"
    echo "$EXTERNAL_IP" > current_ip
    CURRENT_IP=""
fi

if [[ "$CURRENT_IP" == "$EXTERNAL_IP" ]] ; then
    echo "External IP has not changed."
else
    echo "External IP has changed. Setting A record to $EXTERNAL_IP"

    API_URL="https://api.cloudflare.com/client/v4/zones/$ZONE_ID/dns_records"
    HEADER="Authorization: Bearer $API_KEY"

    RES=$(curl -s -H "$HEADER" "$API_URL") 
    DNS_ID=$(echo "$RES" |  jq '.result[] | select(.name == "webdrones.net").id' | tr -d \")

    PATCH_BODY="{\"content\":\"$EXTERNAL_IP\"}"
    API_URL="$API_URL/$DNS_ID"
    echo $API_URL
    curl -s -X PATCH -H "$HEADER" "$API_URL" -d "$PATCH_BODY"
fi

