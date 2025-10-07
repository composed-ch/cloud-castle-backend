#!/usr/bin/bash

curl -v -X POST 'https://backend.cloud-castle.ch/password/reset' -d '{ "email": "patrick.bucher@sluz.ch" }'
