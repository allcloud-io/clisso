#!/bin/bash

if [ "$#" -ne 5 ]; then
    echo "Usage: $0 <access_token> <username_or_email> <password> <app_id> <subdomain>"
    exit 1
fi

access_token=$1
username_or_email=$2
password=$3
app_id=$4
subdomain=$5

response=$(curl -s "https://api.us.onelogin.com/api/1/saml_assertion" \
-X POST \
-H "Authorization: bearer:$access_token" \
-H "Content-Type: application/json" \
-d "{
    \"username_or_email\": \"$username_or_email\",
    \"password\": \"$password\",
    \"app_id\": \"$app_id\",
    \"subdomain\":\"$subdomain\"
}")

device_id=$(echo $response | jq '.data[] .devices[] .device_id' | tr -d '"')
state_token=$(echo $response | jq '.data[] .state_token' | tr -d '"')

# Get OTP token
echo "Please type your OTP token:"
read otp_token

response=$(curl -s "https://api.us.onelogin.com/api/1/saml_assertion/verify_factor" \
-X POST \
-H "Authorization: bearer:$access_token" \
-H "Content-Type: application/json" \
-d "{
    \"app_id\": \"$app_id\",
    \"otp_token\": \"$otp_token\",
    \"device_id\": \"$device_id\",
    \"state_token\": \"$state_token\"
}")

echo $response | jq '.data' | tr -d '"'
