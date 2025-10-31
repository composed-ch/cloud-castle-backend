package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/composed-ch/cloud-castle-backend/internal/config"
)

func main() {
	username := flag.String("username", "", "the unique name of the user")
	zone := flag.String("zone", "", "the zone for which the key is used")
	apiKey := flag.String("key", "", "API key")
	apiSecret := flag.String("secret", "", "API secret")
	tenant := flag.String("tenant", "", "Exoscale tenant (account name)")
	flag.Parse()

	ctx := context.Background()
	conn := config.MustGetConnection()
	defer conn.Close(ctx)

	var accountId uint
	err := conn.QueryRow(ctx, "select id from account where name = $1", username).Scan(&accountId)
	if err != nil {
		fmt.Fprintf(os.Stderr, "selecting account id by username: %v\n", err)
		os.Exit(1)
	}

	_, err = conn.Exec(ctx,
		"insert into api_key (zone, api_key, api_secret, tenant) values ($1, $2, $3, $4)",
		zone, apiKey, apiSecret, tenant)
	if err != nil {
		fmt.Fprintf(os.Stderr, "inserting api key: %v\n", err)
		os.Exit(1)
	}
}
