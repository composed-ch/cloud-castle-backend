# Cloud Castle Backend (in Go)

## Setup

Install goose (for SQL migrations):

    go install github.com/pressly/goose/v3/cmd/goose@latest

## Usage

Register a user:

```sh
go run cmd/register-user/main.go -username joe.doe -role teacher -password topsecret
```

Register an API key for a user:

```sh
go run cmd/add-api-key/main.go -username joe.doe -zone ch-gva-2 -key EXO… -secret SECRET…
```

Login (and store token):

```sh
curl -v -X POST localhost:8080/login -d '{"username": "alice", "password": "topsecret"}' | jq -r '.token' > token.txt
```

Use token:

```sh
curl -v localhost:8080/protected -H "Authorization: Bearer $(cat token.txt)"
```

## Deployment

Create an opearting system user called `cloud_castle`:

    $ sudo useradd -m -d /home/cloud_castle -s $(which bash) cloud_castle

Create the user and database called `cloud_castle`:

    $ sudo -u postgres psql
    =# create user cloud_castle;
    =# create database cloud_castle;
    =# grant all privileges on database cloud_castle to cloud_castle;
    =# \c cloud_castle postgres
    =# grant all on schema public to cloud_castle;

To use the database from the `cloud_castle` user:

    $ sudo -u cloud_castle psql cloud_castle

Create a unit file 

```ini
[Unit]
Description=Cloud Castle Backend
Documentation=https://github.com/composed-ch/cloud-castle-backend-go
After=network.target

[Service]
ExecStart=/home/cloud_castle/bin/cloud_castle
WorkingDirectory=/home/cloud_castle
EnvironmentFile=…
Type=simple
Restart=always

[Install]
WantedBy=multi-user.target
```

## TODO

- for users with the teacher role, show all VMs
- remove username argument from register key script
