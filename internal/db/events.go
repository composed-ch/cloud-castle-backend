package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Kind string

const (
	ACCOUNT_CREATED    Kind = "account_created"
	ACCOUNT_DELETED    Kind = "account_deleted"
	LOGIN_SUCCESS      Kind = "login_success"
	LOGIN_FAILURE      Kind = "login_failure"
	INSTANCE_START     Kind = "instance_start"
	INSTANCE_STOP      Kind = "instance_stop"
	PASSWORD_REQUESTED Kind = "password_requested"
	PASSWORD_RESET     Kind = "password_reset"
)

func LogEvent(ctx context.Context, pool *pgxpool.Pool, kind Kind, accountId int, infoKey, infoVal string) {
	var logWithTimestamp bool = os.Getenv(("LOG_WITH_TIMESTAMP")) == "true"
	var err error
	now := time.Now()
	nowStr := now.Format(time.RFC3339)
	if infoKey != "" || infoVal != "" {
		if logWithTimestamp {
			fmt.Fprintf(os.Stdout, "%s event '%s' by account_id %d (%s=%s)\n", nowStr, kind, accountId, infoKey, infoVal)
		} else {
			fmt.Fprintf(os.Stdout, "event '%s' by account_id %d (%s=%s)\n", kind, accountId, infoKey, infoVal)
		}
		_, err = pool.Exec(ctx, "insert into event_log (account_id, kind, happened, info_key, info_val) values ($1, $2, $3, $4, $5)",
			accountId, kind, now, infoKey, infoVal)
	} else {
		if logWithTimestamp {
			fmt.Fprintf(os.Stdout, "%s event '%s' by account_id %d\n", nowStr, kind, accountId)
		} else {
			fmt.Fprintf(os.Stdout, "event '%s' by account_id %d\n", kind, accountId)
		}
		_, err = pool.Exec(ctx, "insert into event_log (account_id, kind, happened) values ($1, $2, $3)", accountId, kind, now)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "log %s event: %v\n", kind, err)
	}
}
