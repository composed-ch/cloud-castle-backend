package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/composed-ch/cloud-castle-backend/internal/auth"
	"github.com/composed-ch/cloud-castle-backend/internal/config"
	"github.com/composed-ch/cloud-castle-backend/internal/db"
	"go.yaml.in/yaml/v3"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	file := flag.String("file", "", "a group file in YAML format")
	password := flag.String("password", "", "initial password (random if left blank)")
	role := flag.String("role", "student", "user role: 'student' (default) or 'teacher'")
	tenant := flag.String("tenant", "", "Exoscale tenant (account name)")
	flag.Parse()

	if *role != "teacher" && *role != "student" {
		fmt.Fprintf(os.Stderr, "role must either be 'student' or 'teacher'\n")
		os.Exit(1)
	}
	if *tenant == "" {
		fmt.Fprintf(os.Stderr, "missing tenand\n")
		os.Exit(1)
	}

	ctx := context.Background()
	pool := config.MustGetConnectionPool()

	group, err := readGroupFromFile(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read group from file: %v\n", err)
		os.Exit(1)
	}

	for _, user := range group.Users {

		if _, err := db.LoadAccountByName(ctx, pool, user.Name); err == nil {
			fmt.Fprintf(os.Stderr, "user with username '%s' already exists\n", user.Name)
			continue
		}
		var userPassword string
		if *password == "" {
			userPassword, err = auth.RandomPasswordAlnum(32)
			if err != nil {
				fmt.Fprintf(os.Stderr, "generate random password for %v, skipping: %v\n", user, err)
				continue
			}
		} else {
			userPassword = *password
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.DefaultCost)
		if err != nil {
			fmt.Fprintf(os.Stderr, "hash password for user %v, skipping: %v\n", user, err)
			continue
		}
		accountId, err := db.InsertAccount(ctx, pool, user.Name, *role, string(hashedPassword), *tenant, user.Email)
		if err != nil {
			fmt.Fprintf(os.Stderr, "insert user %v: %v\n", user, err)
			continue
		}
		db.LogEvent(ctx, pool, db.ACCOUNT_CREATED, accountId, "name", user.Name)
	}
}

func readGroupFromFile(file string) (*Group, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("open group file: %v", err)
	}
	defer f.Close()
	buf, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read group file: %v", err)
	}
	var group Group
	err = yaml.Unmarshal(buf, &group)
	if err != nil {
		return nil, fmt.Errorf("unmarshal grop file: %v", err)
	}
	return &group, nil
}

type Group struct {
	Name  string `yaml:"name"`
	Users []User `yaml:"users"`
}

type User struct {
	Email  string `yaml:"email"`
	Name   string `yaml:"name"`
	SSHKey string `yaml:"ssh-key"`
}
