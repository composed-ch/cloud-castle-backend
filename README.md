# Cloud Castle Backend (in Go)

## Setup

Install goose (for SQL migrations):

    go install github.com/pressly/goose/v3/cmd/goose@latest

## Usage

Login (and store token):

```sh
curl -v -X POST localhost:8080/login -d '{"username": "alice", "password": "topsecret"}' | jq -r '.token' > token.txt
```

Use token:

```sh
curl -v localhost:8080/protected -H "Authorization: Bearer $(cat token.txt)"
```

## TODO

- [ ] API
    - GET /instances/:id/state
    - GET /instances/:id/start
    - GET /instances/:id/stop