package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/composed-ch/cloud-castle-backend/internal/auth"
	"github.com/composed-ch/cloud-castle-backend/internal/config"
	"github.com/composed-ch/cloud-castle-backend/internal/db"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	username := flag.String("username", "", "the unique name of the user")
	email := flag.String("email", "", "the email address of the user")
	password := flag.String("password", "", "the password used for authentication")
	role := flag.String("role", "student", "user role: 'student' or 'teacher'")
	tenant := flag.String("tenant", "", "Exoscale tenant (account name)")
	flag.Parse()

	conn := config.MustGetDBConecction()
	defer conn.Close(context.Background())

	if _, err := db.LoadAccountByName(conn, *username); err == nil {
		fmt.Fprintf(os.Stderr, "user with username '%s' already exists\n", *username)
		os.Exit(1)
	}

	var userPassword string
	var err error
	if *password == "" {
		userPassword, err = auth.RandomPasswordAlnum(32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "generate random password: %v\n", err)
		}
	} else {
		userPassword = *password
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hash password: %v\n", err)
		os.Exit(1)
	}

	err = db.InsertAccount(conn, *username, *role, string(hashedPassword), *tenant, *email)
	if err != nil {
		fmt.Fprintf(os.Stderr, "insert user %v: %v\n", username, err)
		os.Exit(1)
	}
}
