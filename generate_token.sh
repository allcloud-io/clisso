#!/bin/bash

if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <client_secret> <client_id>"
    exit 1
fi

client_secret=$1
client_id=$2

response=$(curl -s 'https://api.us.onelogin.com/auth/oauth2/token' \
-X POST \
-H "Authorization: client_id:$client_id, client_secret:$client_secret" \
-H "Content-Type: application/json" \
-d '{ 
    "grant_type":"client_credentials"
}')

echo $response | jq '.data[] .access_token' | tr -d '"'
