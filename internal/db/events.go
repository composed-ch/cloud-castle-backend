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
	LOGIN_SUCCESS      Kind = "login_success"
	LOGIN_FAILURE      Kind = "login_failure"
	INSTANCE_START     Kind = "instance_start"
	INSTANCE_STOP      Kind = "instance_stop"
	PASSWORD_REQUESTED Kind = "password_requested"
	PASSWORD_RESET     Kind = "password_reset"
)

func LogEvent(conn *pgx.Conn, ctx context.Context, kind Kind, accountId int, infoKey, infoVal string) error {
	var err error
	now := time.Now()
	nowStr := now.Format(time.RFC3339)
	if infoKey != "" || infoVal != "" {
		fmt.Fprintf(os.Stdout, "%s event '%s' by account_id %d (%s=%s)", nowStr, kind, accountId, infoKey, infoVal)
		_, err = conn.Exec(ctx, "insert into event_log (account_id, kind, happened, info_key, info_val)",
			kind, accountId, now, infoKey, infoVal)
	} else {
		fmt.Fprintf(os.Stdout, "%s event '%s' by account_id %d", nowStr, kind, accountId)
		_, err = conn.Exec(ctx, "insert into event_log (account_id, kind, happened)", kind, accountId, now)
	}
	if err != nil {
		return fmt.Errorf("log %s event: %v", kind, err)
	}
	return nil
}
