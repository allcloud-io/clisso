#!/bin/bash

if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <role_arn> <principal_arn>"
    exit 1
fi

role_arn=$1
principal_arn=$2

aws sts assume-role-with-saml \
    --role-arn $role_arn \
    --principal-arn $principal_arn \
    --saml-assertion $(cat saml.txt)
