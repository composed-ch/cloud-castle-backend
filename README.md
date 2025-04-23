# Cloud Castle Backend (in Go)

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
TODO

- [ ] DB Access & Migration
- [ ] user management with hashed passwords
- [ ] utily to register user
- [ ] store Exoscale token
- [ ] API
    - GET /isntances
    - GET /instances/:id/state
    - GET /instances/:id/start
    - GET /instances/:id/stop