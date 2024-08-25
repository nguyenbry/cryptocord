package crypto

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	headerKey      = "X-CMC_PRO_API_KEY"
	invalidKeyCode = 1001
)

// Error constants
var (
	ErrInvalidKey = errors.New("invalid CoinMarketCap key")
	ErrNotFound   = errors.New("requested token not found")
)

// Non success codes from CoinMarketCap API have this JSON structure
type ErrExternalApi struct {
	Status struct {
		Message string `json:"error_message"`
		Code    int    `json:"error_code"`
	} `json:"status"`
}

// Implement the error interface
func (e *ErrExternalApi) Error() string {
	return fmt.Sprintf("CoinMarketCap error: %s %d", e.Status.Message, e.Status.Code)
}

type Crypto struct {
	cli *http.Client
	key string
}

func New(key string) *Crypto {
	return &Crypto{&http.Client{}, key}
}

func (c *Crypto) buildReq(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add(headerKey, c.key)
	return req, nil
}

func (c *Crypto) decodeErr(res *http.Response) (ErrExternalApi, error) {
	var e ErrExternalApi

	err := json.NewDecoder(res.Body).Decode(&e)

	if err != nil {
		return e, fmt.Errorf("decoding failed request failed: %w", err)
	}

	return e, nil
}

type rawQuote struct {
	Price     float64
	UpdatedAt string
}

// Possible errors are ErrInvalidKey and ErrExternalApi, those returned
// by (json.Decoder).Decode
func (c *Crypto) decodeSuccess(res *http.Response) (rawQuote, error) {
	var out struct {
		Data map[string][](struct {
			Quote map[string]struct {
				Price       float64 `json:"price"`
				LastUpdated string  `json:"last_updated"`
			} `json:"quote"`
			Id int `json:"id"`
		}) `json:"data"`
	}

	err := json.NewDecoder(res.Body).Decode(&out)

	if err != nil {
		return rawQuote{}, fmt.Errorf("decoding success response failed: %w", err)
	}

	btcQuotes, ok := out.Data["BTC"]

	if !ok {
		return rawQuote{}, ErrNotFound
	}

	for _, q := range btcQuotes {
		if q.Id != 1 { // this specific to CoinMarketCap API
			continue
		}

		quote := q.Quote
		p, ok := quote["USD"]

		if !ok {
			return rawQuote{}, ErrNotFound
		}

		return rawQuote{p.Price, p.LastUpdated}, nil
	}

	// not found
	return rawQuote{}, ErrNotFound
}

type quote struct {
	Price     float64
	UpdatedAt time.Time
}

func (c *Crypto) Bitcoin(ctx context.Context) (*quote, error) {
	req, err := c.buildReq(ctx, "https://pro-api.coinmarketcap.com/v2/cryptocurrency/quotes/latest?symbol=BTC%2CUSD")

	if err != nil {
		return nil, fmt.Errorf("building request failed: %w", err)

	}
	res, err := c.cli.Do(req)

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		e, err := c.decodeErr(res)

		if err != nil {
			return nil, err
		}

		if e.Status.Code == invalidKeyCode {
			return nil, ErrInvalidKey
		}

		return nil, &e
	}

	q, err := c.decodeSuccess(res)

	if err != nil {
		return nil, err
	}

	parsedTime, err := time.Parse(time.RFC3339, q.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("parsing time failed: %w", err)
	}

	return &quote{q.Price, parsedTime}, nil

}
