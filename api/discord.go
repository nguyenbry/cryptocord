package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Discord struct {
	url    string
	client *http.Client
}

func NewDiscord(c *http.Client, url string) *Discord {
	return &Discord{url, c}
}

func (d *Discord) Send(c context.Context, text string) error {
	payload := map[string]string{"content": text}
	payloadBytes, err := json.Marshal(payload)

	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(c, "POST", d.url, bytes.NewBuffer(payloadBytes))

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := d.client.Do(req)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		body, err := io.ReadAll(res.Body)

		if err != nil {
			return fmt.Errorf("received non-2xx response and couldn't get body: %d", res.StatusCode)
		}

		return fmt.Errorf("received non-2xx response: %d | (%v)", res.StatusCode, string(body))

	}

	return nil
}
