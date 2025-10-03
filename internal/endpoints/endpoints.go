package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/composed-ch/cloud-castle-backend/exoscale"
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

func (s *Stateful) GetAPIAccess(username string) (*exoscale.APIAccess, error) {
	var zone, key, secret string
	err := s.Pool.QueryRow(context.Background(),
		`select zone, api_key, api_secret
		from api_key
		inner join account on api_key.tenant = account.tenant
		where account.name = $1
		limit 1`, username).Scan(&zone, &key, &secret)
	if err != nil {
		return nil, fmt.Errorf("get API key for %s: %w", username, err)
	}
	return exoscale.NewAPIAccess(username, zone, key, secret), nil
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

func (s *Stateful) GetInstances(w http.ResponseWriter, r *http.Request) {
	api := s.getAPIAccess(w, r)
	if api == nil {
		return
	}
	instances, err := api.GetOwnInstances(api.Username)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get instances: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	payload, err := json.Marshal(instances)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal instances payload: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(payload)
}

func (s *Stateful) GetInstanceState(w http.ResponseWriter, r *http.Request) {
	api := s.getAPIAccess(w, r)
	if api == nil {
		return
	}
	id := r.PathValue("id")
	instance, err := api.GetInstance(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get instance: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	owner, ok := instance.Labels["owner"]
	if !ok || owner != api.Username {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	payload, err := json.Marshal(map[string]string{"state": instance.State})
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal instance payload; %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(payload)
}

func (s *Stateful) StartInstance(w http.ResponseWriter, r *http.Request) {
	api := s.getAPIAccess(w, r)
	if api == nil {
		return
	}
	id := r.PathValue("id")
	instance, err := api.GetInstance(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get instance: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	owner, ok := instance.Labels["owner"]
	if !ok || owner != api.Username {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	err = api.StartInstance(id)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(200)
}

func (s *Stateful) StopInstance(w http.ResponseWriter, r *http.Request) {
	api := s.getAPIAccess(w, r)
	if api == nil {
		return
	}
	id := r.PathValue("id")
	instance, err := api.GetInstance(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get instance: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	owner, ok := instance.Labels["owner"]
	if !ok || owner != api.Username {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	err = api.StopInstance(id)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(200)
}

func (s *Stateful) getAPIAccess(w http.ResponseWriter, r *http.Request) *exoscale.APIAccess {
	authorization := r.Header.Get("Authorization")
	subject, err := auth.ExtractSubject(authorization)
	if err != nil {
		fmt.Fprintf(os.Stderr, "extract subject: %v\n", err)
		w.WriteHeader(http.StatusUnauthorized)
		return nil
	}
	api, err := s.GetAPIAccess(subject)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get API access: %v\n", err)
		w.WriteHeader(http.StatusUnauthorized)
		return nil
	}
	return api
}
