package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/composed-ch/cloud-castle-backend/internal/auth"
	"github.com/composed-ch/cloud-castle-backend/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Stateful struct {
	Pool *pgxpool.Pool
}

func NewStateful(cfg *config.Config) (*Stateful, error) {
	pool, err := pgxpool.New(context.Background(), cfg.BuildDatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}
	return &Stateful{Pool: pool}, nil
}

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
}

func (s *Stateful) Login(w http.ResponseWriter, r *http.Request) {
	var authPayload authRequest
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	if err = json.Unmarshal(payload, &authPayload); err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	var hashed string
	err = s.Pool.QueryRow(
		context.Background(),
		"select password from account where name = $1",
		authPayload.Username).Scan(&hashed)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte(authPayload.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	tokenStr, err := auth.IssueToken(authPayload.Username)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if tokenData, err := json.Marshal(authResponse{Token: tokenStr}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Write(tokenData)
	}
}

func Blah(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Greetings, Sire!\n"))
}
