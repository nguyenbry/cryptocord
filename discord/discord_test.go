package discord

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

func TestSendMessageSuceeds(t *testing.T) {
	d := New(&http.Client{})

	err := d.Send(context.Background(), testUrl, "hey")

	if err != nil {
		t.Fatalf("sending message to webhook failed: %v", err)
	}
}

func TestExpiredContextFails(t *testing.T) {
	d := New(&http.Client{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := d.Send(ctx, testUrl, "hey")

	if err == nil {
		t.Fatal("expected request to fail due to expired context")
	}
}
