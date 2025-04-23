package endpoints

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/composed-ch/cloud-castle-backend/internal/auth"
)

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

var (
	logins map[string]string = map[string]string{
		"alice": "topsecret",
		"bob":   "mossecret",
	}
)

func Login(w http.ResponseWriter, r *http.Request) {
	var authPayload AuthRequest
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	if err = json.Unmarshal(payload, &authPayload); err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	password, ok := logins[authPayload.Username]
	if !ok || password != authPayload.Password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	tokenStr, err := auth.IssueToken(authPayload.Username)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if tokenData, err := json.Marshal(AuthResponse{Token: tokenStr}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Write(tokenData)
	}
}

func Blah(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Greetings, Sire!\n"))
}
