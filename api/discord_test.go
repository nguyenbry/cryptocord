package main

import (
	"context"
	"net/http"
	"os"
	"testing"
)

var testUrl string

func init() {
	url, ok := os.LookupEnv("WH")

	if !ok {
		panic("test webhook url must be provided")
	}

	testUrl = url
}

func TestSendMessage(t *testing.T) {
	d := NewDiscord(&http.Client{}, testUrl)

	err := d.Send(context.Background(), "hey")

	if err != nil {
		t.Fatalf("sending message to webhook failed: %v", err)
	}
}
