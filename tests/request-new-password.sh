#!/usr/bin/bash

curl -v -X POST http://localhost:8080/password/reset -d '{ "email": "patrick.bucher@composed.ch" }'