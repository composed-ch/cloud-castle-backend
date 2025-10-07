curl -v -X POST http://localhost:8080/password/new \
    -d '{ "email": "patrick.bucher@composed.ch", "token": "superSecretTokenGeneratedAndSentByEmail", "password": "123456789", "confirmation": "123456789" }'

