package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const secret = "utmostsecret"

var signingMethod *jwt.SigningMethodHMAC = jwt.SigningMethodHS512

var authHeaderPattern = regexp.MustCompilePOSIX("^Bearer (.+)$")

var logins map[string]string = map[string]string{
	"alice": "topsecret",
	"bob":   "mossecret",
}

type AuthReques struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /login", login)
	mux.HandleFunc("GET /protected", auth(blah))
	http.ListenAndServe("localhost:8080", mux)
}

type Handler func(http.ResponseWriter, *http.Request)

func auth(handler Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		authorization := strings.TrimSpace(r.Header.Get("Authorization"))
		matches := authHeaderPattern.FindStringSubmatch(authorization)
		if len(matches) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		tokenStr := matches[1]
		if _, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{signingMethod.Alg()})); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			handler(w, r)
		}
	}
}

func blah(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Greetings, Sire!"))
}

func login(w http.ResponseWriter, r *http.Request) {
	var authPayload AuthReques
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
		now := time.Now()
		exp := now.Add(time.Hour * 24)
		token := jwt.NewWithClaims(signingMethod, jwt.MapClaims{
			"sub": authPayload.Username,
			"iat": now.Unix(),
			"exp": exp.Unix(),
		})
		if tokenStr, err := token.SignedString([]byte(secret)); err != nil {
			fmt.Fprintln(os.Stderr, err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			if tokenData, err := json.Marshal(AuthResponse{Token: tokenStr}); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Write(tokenData)
			}
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}
