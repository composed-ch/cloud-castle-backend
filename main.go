package main

import (
	"encoding/json"
	"io"
	"net/http"
)

var logins map[string]string = map[string]string{
	"alice": "topsecret",
	"bob":   "mossecret",
}

type AuthPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /login", login)
	http.ListenAndServe("localhost:8080", mux)
}

func login(w http.ResponseWriter, r *http.Request) {
	var authPayload AuthPayload
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	if err = json.Unmarshal(payload, &authPayload); err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	if password, ok := logins[authPayload.Username]; ok && password == authPayload.Password {
		w.Write([]byte("ok"))
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}
