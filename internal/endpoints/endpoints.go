package endpoints

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/composed-ch/cloud-castle-backend/exoscale"
	"github.com/composed-ch/cloud-castle-backend/internal/auth"
	"github.com/composed-ch/cloud-castle-backend/internal/config"
	"github.com/composed-ch/cloud-castle-backend/internal/mailing"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Stateful struct {
	Pool   *pgxpool.Pool
	Config *config.Config
}

func NewStateful(cfg *config.Config) (*Stateful, error) {
	pool, err := pgxpool.New(context.Background(), cfg.BuildDatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}
	return &Stateful{Pool: pool, Config: cfg}, nil
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
		fmt.Fprintf(os.Stderr, "login attempt for user %s failed: %v\n", authPayload.Username, err)
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
		fmt.Fprintf(os.Stderr, "successful login attempt for user %s\n", authPayload.Username)
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
	fmt.Fprintf(os.Stderr, "user %s started instance %s\n", owner, id)
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
	fmt.Fprintf(os.Stderr, "user %s stopped instance %s\n", owner, id)
	w.WriteHeader(200)
}

func (s *Stateful) ResetPassword(w http.ResponseWriter, r *http.Request) {
	type Payload struct {
		Email string `json:"email"`
	}
	var payload *Payload
	payload, err := jsonBody[Payload](r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unmarshal password reset request body: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var accountId int
	if err := s.Pool.QueryRow(context.Background(), "select id from account where email = $1", payload.Email).Scan(&accountId); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			fmt.Fprintf(os.Stderr, "no account found for password reset by email %s: %v\n", payload.Email, err)
			w.WriteHeader(http.StatusCreated)
		} else {
			fmt.Fprintf(os.Stderr, "fetch account for password reset by email %s: %v\n", payload.Email, err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	var created sql.NullTime
	hasExistingRequest := false
	err = s.Pool.QueryRow(context.Background(), "select max(created) from password_reset where account_id = $1", accountId).Scan(&created)
	if err == nil && created.Valid && created.Time.Add(time.Minute*5).After(time.Now()) {
		fmt.Fprintf(os.Stderr, "password reset request coming in too soon for %s\n", payload.Email)
		w.WriteHeader(http.StatusTooManyRequests)
		return
	} else if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			fmt.Fprintf(os.Stderr, "query for existing password reset requests for %s: %v\n", payload.Email, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	token, err := auth.RandomPasswordAlnum(64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate random password: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	hashedToken, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bcrypt token: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if hasExistingRequest {
		if _, err := s.Pool.Exec(context.Background(), "update password_reset set token = $1 where account_id = $2", hashedToken, accountId); err != nil {
			fmt.Fprintf(os.Stderr, "update passwort reset token: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		if _, err := s.Pool.Exec(context.Background(), "insert into password_reset (account_id, token) values ($1, $2)", accountId, hashedToken); err != nil {
			fmt.Fprintf(os.Stderr, "insert passwort reset token: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	message := mailing.CreatePasswordResetEmail(accountId, payload.Email, token)
	err = mailing.SendPostmarkEmail(
		"info@cloud-castle.ch",
		payload.Email,
		"Cloud Castle Password Reset",
		message,
		"cloud-castle-password-reset",
		s.Config.PostmarkToken,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "send password reset email to %s: %v\n", payload.Email, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(os.Stderr, "sent password reset email to %s\n", payload.Email)
}
func (s *Stateful) NewPassword(w http.ResponseWriter, r *http.Request) {
	type Payload struct {
		Email    string `json:"email"`
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	payload, err := jsonBody[Payload](r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unmarshal new password request body: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !auth.SufficientlyStrong(payload.Password) {
		fmt.Fprintf(os.Stderr, "provided password '%s' is too weak by our standards, rejecting reset\n", payload.Password)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var accountId int
	err = s.Pool.QueryRow(context.Background(), "select id from account where email = $1", payload.Email).Scan(&accountId)
	if err != nil {
		fmt.Fprintf(os.Stderr, "no account found for email address %s: %v\n", payload.Email, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var expires sql.NullTime
	var token sql.NullString
	var passwordResetId sql.NullInt64
	err = s.Pool.QueryRow(context.Background(), "select id, token, expires from password_reset where account_id = $1", accountId).Scan(&passwordResetId, &token, &expires)
	if err != nil || !expires.Valid || !token.Valid || !passwordResetId.Valid {
		fmt.Fprintf(os.Stderr, "no password reset request found for accountId %d: %v\n", accountId, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if expires.Time.Before(time.Now()) {
		fmt.Fprintf(os.Stderr, "the password reset request with id %d expired at %v\n", passwordResetId.Int64, expires.Time)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err = bcrypt.CompareHashAndPassword([]byte(token.String), []byte(payload.Token)); err != nil {
		fmt.Fprintf(os.Stderr, "the token provided does not match the one from the database (id: %d): %v", passwordResetId.Int64, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hashing password %s: %v\n", payload.Password, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err := s.Pool.Exec(context.Background(), "update account set password = $1 where id = $2", hashedPassword, accountId); err != nil {
		fmt.Fprintf(os.Stderr, "update password for account %d: %v\n", accountId, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err := s.Pool.Exec(context.Background(), "delete from password_reset where account_id = $1", accountId); err != nil {
		fmt.Fprintf(os.Stderr, "deleting password reset entries for accountId %d: %v", accountId, err)
	}
	fmt.Fprintf(os.Stderr, "updated password for accountId %d\n", accountId)
	w.WriteHeader(http.StatusNoContent)
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

func jsonBody[T any](r *http.Request) (*T, error) {
	var payload T
	buf := bytes.NewBufferString("")
	_, err := io.Copy(buf, r.Body)
	if err != nil {
		return nil, fmt.Errorf("copy body: %v", err)
	}
	err = json.Unmarshal(buf.Bytes(), &payload)
	if err != nil {
		return nil, fmt.Errorf("unmarshal %s: %v", buf.String(), err)
	}
	return &payload, nil
}
