package mailing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// TODO: consider proper HTML template
// TODO: read frontend url from environment for proper link
func CreatePasswordResetEmail(accountId int, email, token string) string {
	atIndex := strings.Index(email, "@")
	username := email[:atIndex]
	resetURL := fmt.Sprintf("https://app.cloud-castle.ch/password-reset/%s", token)
	return fmt.Sprintf(
		`<p>Hallo %s!</p>
		<p>Du hast ein neues Password für <a href="https://app.cloud-castle.ch">Cloud Castle</a> angefragt.</p>
		<p>Wenn du das nicht warst, kannst du diese Nachricht löschen.</p>
		<p>Wenn du das warst, kannst du ein <a href="%s">dein Passwort zurücksetzen</a>.</p>
		<p>Liebe Grüsse vom Cloud Castle!</p>`, username, resetURL)
}

type Email struct {
	From          string
	To            string
	Subject       string
	HTMLBody      string `json:"HtmlBody"`
	MessageStream string
}

func SendPostmarkEmail(from, to, subject, content, stream, postmarkToken string) error {
	body := Email{
		From:          from,
		To:            to,
		Subject:       subject,
		HTMLBody:      content,
		MessageStream: stream,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal message body: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, "https://api.postmarkapp.com/email", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("prepare request for postmark: %v", err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Postmark-Server-Token", postmarkToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send email via postmark: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("send email via postmark: status code: %d", res.StatusCode)
	}
	return nil
}
