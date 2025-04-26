package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/composed-ch/cloud-castle-backend/internal/auth"
	"github.com/composed-ch/cloud-castle-backend/internal/config"
	"github.com/composed-ch/cloud-castle-backend/internal/endpoints"
	"github.com/composed-ch/cloud-castle-backend/internal/middleware"
)

func main() {
	cfg := config.MustReadConfig()
	state, err := endpoints.NewStateful(&cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "initializing state: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /login", state.Login)
	mux.HandleFunc("GET /protected", auth.Authenticated(endpoints.Blah))
	http.ListenAndServe("localhost:8080", middleware.AllowCORS(mux))
}
