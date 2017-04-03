#!/bin/bash

#if [ "$#" -ne 2 ]; then
#    echo "Usage: $0 <client_secret> <client_id>"
#    exit 1
#fi

role_arn=$1
principal_arn=$2

aws sts assume-role-with-saml \
    --role-arn $role_arn \
    --principal-arn $principal_arn \
    --saml-assertion $(cat saml.txt)
