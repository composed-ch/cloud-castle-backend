package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
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

func LogEvent(conn *pgx.Conn, ctx context.Context, kind Kind, accountId int, infoKey, infoVal string) {
	var err error
	now := time.Now()
	nowStr := now.Format(time.RFC3339)
	if infoKey != "" || infoVal != "" {
		fmt.Fprintf(os.Stdout, "%s event '%s' by account_id %d (%s=%s)\n", nowStr, kind, accountId, infoKey, infoVal)
		_, err = conn.Exec(ctx, "insert into event_log (account_id, kind, happened, info_key, info_val) values ($1, $2, $3, $4, $5)",
			accountId, kind, now, infoKey, infoVal)
	} else {
		fmt.Fprintf(os.Stdout, "%s event '%s' by account_id %d\n", nowStr, kind, accountId)
		_, err = conn.Exec(ctx, "insert into event_log (account_id, kind, happened) values ($1, $2, $3)", accountId, kind, now)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "log %s event: %v\n", kind, err)
	}
}
