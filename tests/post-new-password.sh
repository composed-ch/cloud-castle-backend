#!/usr/bin/bash

curl -v -X POST 'https://backend.cloud-castle.ch/password/new' \
    -d '{ "email": "patrick.bucher@sluz.ch", "token": "", "password": "topsecret" }'
