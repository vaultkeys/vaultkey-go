package vaultkey

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	mainnetBaseURL = "https://app.vaultkeys.dev/api/v1/sdk"
	testnetBaseURL = "https://testnet.vaultkeys.dev/api/v1/sdk"
	userAgent      = "vaultkey-go"
)

// Client is the VaultKey API client.
type Client struct {
	apiKey     string
	apiSecret  string
	baseURL    string
	httpClient *http.Client

	Wallets     *WalletsService
	Jobs        *JobsService
	Stablecoin  *StablecoinService
	Chains      *ChainsService
}

// ClientOption configures the Client.
type ClientOption func(*Client)

// WithBaseURL overrides the base API URL.
// Useful for self-hosted deployments or proxies.
// When set, this takes precedence over automatic key-based routing.
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithHTTPClient provides a custom HTTP client.
func WithHTTPClient(h *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = h
	}
}

// resolveBaseURL picks the correct endpoint from the API key prefix.
// testnet_ keys → https://testnet.vaultkeys.dev
// vk_live_ keys → https://app.vaultkeys.dev
func resolveBaseURL(apiKey string) string {
	if strings.HasPrefix(apiKey, "testnet_") {
		return testnetBaseURL
	}
	return mainnetBaseURL
}

// NewClient creates a new VaultKey API client.
//
// apiKey and apiSecret are required. If either is empty, the client falls back
// to the VAULTKEY_API_KEY and VAULTKEY_API_SECRET environment variables.
//
// The correct endpoint (testnet or mainnet) is selected automatically based on
// the api key prefix. Use WithBaseURL to override this for self-hosted deployments.
func NewClient(apiKey, apiSecret string, opts ...ClientOption) (*Client, error) {
	if apiKey == "" {
		apiKey = os.Getenv("VAULTKEY_API_KEY")
	}
	if apiSecret == "" {
		apiSecret = os.Getenv("VAULTKEY_API_SECRET")
	}
	if apiKey == "" {
		return nil, errors.New("missing API key: pass it to NewClient or set VAULTKEY_API_KEY")
	}
	if apiSecret == "" {
		return nil, errors.New("missing API secret: pass it to NewClient or set VAULTKEY_API_SECRET")
	}

	c := &Client{
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		baseURL:    resolveBaseURL(apiKey),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	for _, opt := range opts {
		opt(c)
	}

	c.Wallets = &WalletsService{client: c}
	c.Jobs = &JobsService{client: c}
	c.Stablecoin = &StablecoinService{client: c}
	c.Chains = &ChainsService{client: c}

	return c, nil
}

func (c *Client) doRequest(ctx context.Context, method, path string, body any, v any) (*ErrorResponse, error) {
	var buf io.Reader
	if body != nil {
		b := &bytes.Buffer{}
		if err := json.NewEncoder(b).Encode(body); err != nil {
			return nil, err
		}
		buf = b
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("X-API-Secret", c.apiSecret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if v != nil {
			if err := json.NewDecoder(resp.Body).Decode(v); err != nil && err != io.EOF {
				return nil, err
			}
		}
		return nil, nil
	}

	// Attempt to decode a structured error body.
	// Fall back to a generic error if decoding fails.
	errResp := &ErrorResponse{Message: resp.Status, Code: "INTERNAL_SERVER_ERROR"}
	_ = json.NewDecoder(resp.Body).Decode(errResp)
	return errResp, nil
}

func (c *Client) get(ctx context.Context, path string, out any) (*ErrorResponse, error) {
	return c.doRequest(ctx, http.MethodGet, path, nil, out)
}

func (c *Client) post(ctx context.Context, path string, body any, out any) (*ErrorResponse, error) {
	return c.doRequest(ctx, http.MethodPost, path, body, out)
}

func (c *Client) delete(ctx context.Context, path string, body any, out any) (*ErrorResponse, error) {
	return c.doRequest(ctx, http.MethodDelete, path, body, out)
}