package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/composed-ch/cloud-castle-backend/internal/auth"
	"github.com/composed-ch/cloud-castle-backend/internal/config"
	"github.com/composed-ch/cloud-castle-backend/internal/db"
)

func main() {
	name := flag.String("name", "", "the name of the account for which the password shall be set")
	password := flag.String("password", "", "the password that shall be set for the account")
	flag.Parse()

	if *name == "" || *password == "" {
		fmt.Fprintf(os.Stderr, "both name and password are requird\n")
		os.Exit(1)
	}

	ctx := context.Background()
	pool := config.MustGetConnectionPool()

	account, err := db.LoadAccountByName(ctx, pool, *name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load account by name '%s': %v\n", *name, err)
		os.Exit(1)
	}

	if !auth.SufficientlyStrong(*password) {
		fmt.Fprintf(os.Stderr, "the given password is too weak\n")
		os.Exit(1)
	}

	if err = db.UpdatePassword(ctx, pool, account.Name, *password); err != nil {
		fmt.Fprintf(os.Stderr, "update password for account '%s': %v\n", account.Name, err)
	}
}
