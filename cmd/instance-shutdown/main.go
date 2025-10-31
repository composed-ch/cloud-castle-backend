package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/composed-ch/cloud-castle-backend/internal/config"
	"github.com/composed-ch/cloud-castle-backend/internal/db"
	"github.com/composed-ch/cloud-castle-backend/internal/endpoints"
)

func main() {
	label := flag.String("label", "", "the label to select the instances (all if empty)")
	value := flag.String("value", "", "the label value to select the instances (all if empty)")
	user := flag.String("user", "", "the user to hold accountable for the shutdown (determines the tenant)")
	flag.Parse()

	if *label != "" && *value == "" || *label == "" && *value != "" {
		fmt.Fprintln(os.Stderr, "define both label and value or neither")
		os.Exit(1)
	}

	cfg := config.MustReadConfig()
	state, err := endpoints.NewStateful(&cfg)
	defer state.Pool.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "initializing state: %v\n", err)
		os.Exit(1)
	}

	api, err := state.GetAPIAccess(*user)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get API access for user '%s': %v\n", *user, err)
		os.Exit(1)
	}

	ctx := context.Background()
	accountId, err := db.LoadAccountIdByName(ctx, state.Pool, *user)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load accountId by name '%s': %v\n", *user, err)
		os.Exit(1)
	}

	instances, err := api.GetInstances()
	if err != nil {
		fmt.Fprintf(os.Stderr, "get instances: %v\n", err)
		os.Exit(1)
	}

	shutdownInstanceIds := make([]string, 0)
	for _, instance := range instances {
		if instance.State != "running" {
			continue
		}
		if *label != "" {
			if instanceValue, ok := instance.Labels[*label]; ok && instanceValue == *value {
				shutdownInstanceIds = append(shutdownInstanceIds, instance.ID)
			}
		} else {
			shutdownInstanceIds = append(shutdownInstanceIds, instance.ID)
		}
	}

	for _, id := range shutdownInstanceIds {
		if err := api.StopInstance(id); err != nil {
			fmt.Fprintf(os.Stderr, "shutdown instance %s: %v\n", id, err)
		} else {
			db.LogEvent(ctx, state.Pool, db.INSTANCE_STOP, accountId, "instance", id)
		}
	}
}
