package main

import (
	"net/http"

	"github.com/composed-ch/cloud-castle-backend/internal/auth"
	"github.com/composed-ch/cloud-castle-backend/internal/endpoints"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /login", endpoints.Login)
	mux.HandleFunc("GET /protected", auth.Authenticated(endpoints.Blah))
	http.ListenAndServe("localhost:8080", mux)
}
