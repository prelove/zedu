package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
)

type ResendSender struct {
	key, from string
	client    *http.Client
}

func NewResendSenderFromEnv() (*ResendSender, error) {
	key, from := os.Getenv("ZEDU_RESEND_API_KEY"), os.Getenv("ZEDU_RESEND_FROM")
	if key == "" || from == "" {
		return nil, errors.New("resend configuration missing")
	}
	return &ResendSender{key: key, from: from, client: http.DefaultClient}, nil
}
func (s *ResendSender) Send(ctx context.Context, to, subject, body string) (string, error) {
	payload, _ := json.Marshal(map[string]any{"from": s.from, "to": []string{to}, "subject": subject, "html": body})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.key)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.New("resend rejected notification")
	}
	var result struct {
		ID string `json:"id"`
	}
	if json.NewDecoder(resp.Body).Decode(&result) != nil {
		return "", errors.New("resend response invalid")
	}
	return result.ID, nil
}
