package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/composed-ch/cloud-castle-backend/internal/config"
	"github.com/composed-ch/cloud-castle-backend/internal/endpoints"
)

func main() {
	username := flag.String("username", "", "the unique name of the user")
	zone := flag.String("zone", "", "the zone for which the key is used")
	apiKey := flag.String("key", "", "API key")
	apiSecret := flag.String("secret", "", "API secret")
	tenant := flag.String("tenant", "", "Exoscale tenant (account name)")
	flag.Parse()

	cfg := config.MustReadConfig()
	state, err := endpoints.NewStateful(&cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "initializing state: %v\n", err)
		os.Exit(1)
	}

	var accountId uint
	err = state.Pool.QueryRow(
		context.Background(),
		"select id from account where name = $1",
		username).Scan(&accountId)
	if err != nil {
		fmt.Fprintf(os.Stderr, "selecting account id by username: %v\n", err)
		os.Exit(1)
	}

	_, err = state.Pool.Exec(context.Background(),
		"insert into api_key (zone, api_key, api_secret, tenant) values ($1, $2, $3, $4)",
		zone, apiKey, apiSecret, tenant)
	if err != nil {
		fmt.Fprintf(os.Stderr, "inserting api key: %v\n", err)
		os.Exit(1)
	}
}
