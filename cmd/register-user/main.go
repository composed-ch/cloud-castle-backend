package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/composed-ch/cloud-castle-backend/internal/config"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	username := flag.String("username", "", "the unique name of the user")
	password := flag.String("password", "", "the password used for authentication")
	role := flag.String("role", "student", "user role: 'student' or 'teacher'")
	flag.Parse()

	cfg := config.MustReadConfig()
	url := cfg.BuildDatabaseURL()
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect to database: %s", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	hashed, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hash password: %v", err)
	}

	_, err = conn.Exec(context.Background(),
		"insert into account (name, role, password) values ($1, $2, $3)",
		username, role, hashed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "insert user: %v", err)
		os.Exit(1)
	}
}
