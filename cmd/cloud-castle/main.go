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
		fmt.Fprintf(os.Stderr, "initializing state: %v\n", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /canary", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("OK")) })
	mux.HandleFunc("POST /login", state.Login)
	mux.HandleFunc("GET /instances", auth.Authenticated(state.GetInstances))
	mux.HandleFunc("GET /instance/{id}/state", auth.Authenticated(state.GetInstanceState))
	mux.HandleFunc("GET /instance/{id}/start", auth.Authenticated(state.StartInstance))
	mux.HandleFunc("GET /instance/{id}/stop", auth.Authenticated(state.StopInstance))
	http.ListenAndServe("localhost:8080", middleware.AllowCORS(mux))
}
