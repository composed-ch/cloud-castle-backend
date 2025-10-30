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
	Id         int
	Name       string
	Role       string
	Registered time.Time
	Password   string
	Tenant     string
	Email      string
}

// TODO: add proper context
func InsertAccount(conn *pgx.Conn, name, role, hashedPassword, tenant, email string) error {
	_, err := conn.Exec(context.TODO(),
		"insert into account (name, role, password, tenant, email) values ($1, $2, $3, $4, $5)",
		name, role, hashedPassword, tenant, email)
	if err != nil {
		return fmt.Errorf("insert user: %v", err)
	}
	return nil
}

func LoadAccountIdByName(ctx context.Context, conn *pgx.Conn, name string) (int, error) {
	var id int
	if err := conn.QueryRow(ctx, "select id from account where name = $1", name).Scan(&id); err != nil {
		return -1, fmt.Errorf("load account id by name '%s': %v", name, err)
	}
	return id, nil
}

func LoadAccountIdByEmail(ctx context.Context, conn *pgx.Conn, email string) (int, error) {
	var id int
	if err := conn.QueryRow(ctx, "select id from account where email = $1", email).Scan(&id); err != nil {
		return -1, fmt.Errorf("load account id by name '%s': %v", email, err)
	}
	return id, nil
}

// TODO: add proper context
func LoadAccountByName(conn *pgx.Conn, name string) (*Account, error) {
	var registered sql.NullTime
	var role, password, tenant, email sql.NullString
	var id sql.NullInt32
	err := conn.QueryRow(context.TODO(),
		"select id, role, registered, password, tenant, email from account where name = $1",
		name).Scan(&id, &role, &registered, &password, &tenant, &email)
	if err != nil {
		return nil, fmt.Errorf(`load account by name "%s": %v`, name, err)
	}
	return &Account{
		Id:         int(id.Int32),
		Name:       name,
		Role:       role.String,
		Registered: registered.Time,
		Password:   password.String,
		Tenant:     tenant.String,
		Email:      email.String,
	}, nil
}

// TODO: add proper context
func UpdatePassword(conn *pgx.Conn, name, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %v", err)
	}
	_, err = conn.Exec(context.TODO(), "update account set password = $1 where name = $2", hashedPassword, name)
	if err != nil {
		return fmt.Errorf("update account with name '%s': %v", name, err)
	}
	return nil
}
