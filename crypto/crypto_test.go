package crypto

import (
	"context"
	"errors"
	"os"
	"testing"
)

func TestHttpClientNonNil(t *testing.T) {
	c := New("test key")

	if c.cli == nil {
		t.Fatalf("expected non-nil client")
	}
}

func TestBuildReqNeverFails(t *testing.T) {
	c := New("test key")

	req, err := c.buildReq(context.Background(), "https://example.com")

	if err != nil {
		t.Fatalf("creating request should never fail: %v", err)
		return
	}

	if req == nil {
		t.Fatalf("must be non-nil request")
	}
}

func TestKeyAddedToHeader(t *testing.T) {
	testKey := "test key"
	req, _ := New(testKey).buildReq(context.Background(), "https://example.com")
	header := req.Header

	if header == nil {
		t.Fatalf("must be non-nil header")
	}

	if header.Get(headerKey) != testKey {
		t.Errorf("expected key to be added to header")
	}
}

func TestReqWithBadKeyReturnError(t *testing.T) {
	c := New("bad key")
	_, err := c.Bitcoin(context.Background())

	if err == nil {
		t.Fatalf("expected error with bad key. perhaps a 403")
		return
	}

	if err != ErrInvalidKey {
		t.Fatalf("expected ErrInvalidKey error but got: %v", err)
		return
	}
}

func TestSuccessfulReq(t *testing.T) {
	key, exists := os.LookupEnv("CMC_KEY")

	if !exists {
		t.Fatalf("CMC_KEY must be set in env")
	}

	c := New(key)

	quote, err := c.Bitcoin(context.Background())

	if err != nil {
		if errors.Is(err, ErrInvalidKey) {
			t.Fatalf("key provided was rejected by CoinMarketCap")
		}

		if quote.Price != 0 {
			t.Fatalf("expected price to be 0")
		}
		t.Fatalf("expected no error but got: %v", err)
	}

	if quote.Price <= 0 {
		t.Fatalf("expected price to be greater than 0")
	}
}
