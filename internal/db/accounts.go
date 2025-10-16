package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	Name       string
	Role       string
	Registered time.Time
	Password   string
	Tenant     string
	Email      string
}

func InsertAccount(conn *pgx.Conn, name, role, hashedPassword, tenant, email string) error {
	_, err := conn.Exec(context.Background(),
		"insert into account (name, role, password, tenant, email) values ($1, $2, $3, $4, $5)",
		name, role, hashedPassword, tenant, email)
	if err != nil {
		return fmt.Errorf("insert user: %v", err)
	}
	return nil
}

func LoadAccountByName(conn *pgx.Conn, name string) (*Account, error) {
	var registered sql.NullTime
	var role, password, tenant, email sql.NullString
	err := conn.QueryRow(context.Background(), "select role, registered, password, tenant, email from account where name = $1", name).Scan(&role, &registered, &password, &tenant, &email)
	if err != nil {
		return nil, fmt.Errorf(`load account by name "%s": %v`, name, err)
	}
	return &Account{
		Name:       name,
		Role:       role.String,
		Registered: registered.Time,
		Password:   password.String,
		Tenant:     tenant.String,
		Email:      email.String,
	}, nil
}

func UpdatePassword(conn *pgx.Conn, name, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %v", err)
	}
	_, err = conn.Exec(context.Background(), "update account set password = $1 where name = $2", hashedPassword, name)
	if err != nil {
		return fmt.Errorf("update account with name '%s': %v", name, err)
	}
	return nil
}
