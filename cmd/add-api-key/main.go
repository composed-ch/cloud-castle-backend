package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/composed-ch/cloud-castle-backend/internal/config"
	"github.com/jackc/pgx/v5"
)

func main() {
	username := flag.String("username", "", "the unique name of the user")
	zone := flag.String("zone", "", "the zone for which the key is used")
	apiKey := flag.String("key", "", "API key")
	apiSecret := flag.String("secret", "", "API secret")
	tenant := flag.String("tenant", "", "Exoscale tenant (account name)")
	flag.Parse()

	cfg := config.MustReadConfig()
	url := cfg.BuildDatabaseURL()
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect to database: %s", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	var accountId uint
	err = conn.QueryRow(
		context.Background(),
		"select id from account where name = $1",
		username).Scan(&accountId)
	if err != nil {
		fmt.Fprintf(os.Stderr, "selecting account id by username: %v", err)
		os.Exit(1)
	}

	_, err = conn.Exec(context.Background(),
		"insert into api_key (zone, api_key, api_secret, tenant) values ($1, $2, $3, $4)",
		zone, apiKey, apiSecret, tenant)
	if err != nil {
		fmt.Fprintf(os.Stderr, "inserting api key: %v", err)
		os.Exit(1)
	}
}
