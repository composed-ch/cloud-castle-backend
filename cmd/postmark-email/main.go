package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type Email struct {
	From          string
	To            string
	Subject       string
	HTMLBody      string `json:"HtmlBody"`
	MessageStream string
}

func main() {
	body := Email{
		From:          "patrick.bucher@composed.ch",
		To:            "patrick.bucher@composed.ch",
		Subject:       "Cloud Castle: Dein neues Passwort",
		HTMLBody:      "Klopfe dreimal an die Zugbrücke um Einlass zu erhalten.",
		MessageStream: "cloud-castle-password-reset",
	}
	payload, err := json.Marshal(body)
	if err != nil {
		log.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, "https://api.postmarkapp.com/email", bytes.NewBuffer(payload))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Postmark-Server-Token", os.Getenv("POSTMARK_TOKEN"))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res.StatusCode)
}
