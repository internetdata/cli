package main

import (
	"errors"
	"fmt"

	internetdata "github.com/internetdata/sdk-go/v2"
)

// Client is the SDK client the current invocation uses, and the key it holds.
type Client struct {
	api *internetdata.Client
	key string
}

// NewClient builds the client the current invocation should use.
func NewClient() (*Client, error) {
	key, _ := gConfig.ResolveKey()

	opts := []internetdata.Option{
		internetdata.WithRetries(resolveRetries()),
	}
	if key != "" {
		opts = append(opts, internetdata.WithAPIKey(key))
	}
	baseURL := gConfig.ResolveBaseURL()
	if baseURL != "" {
		opts = append(opts, internetdata.WithBaseURL(baseURL))
	}

	api, err := internetdata.New(opts...)
	if err != nil {
		return nil, err
	}
	return &Client{api: api, key: key}, nil
}

// HasKey reports whether this invocation is authenticated.
func (c *Client) HasKey() bool { return c.key != "" }

// Database is the licensed-database half of the API.
func (c *Client) Database() *internetdata.DatabaseAPI { return c.api.Database }

// requireKey refuses a command that cannot work unauthenticated, naming how to
// fix it rather than reporting a 401 from three layers down.
func (c *Client) requireKey(what string) error {
	if c.HasKey() {
		return nil
	}
	return fmt.Errorf(
		"%s needs an API key; run `%s login`, or pass --key, or set INTERNETDATA_API_KEY",
		what, progBase,
	)
}

// resolveRetries is how many times a retryable failure is retried.
func resolveRetries() int {
	if fRetries >= 0 {
		return fRetries
	}
	if gConfig.Retries >= 0 {
		return gConfig.Retries
	}
	return defaultRetries
}

// explain turns an SDK error into something worth reading at a terminal.
//
// The SDK's own message is accurate and says nothing about what to DO, which
// for the two authentication failures is the entire question. Keys are
// default-deny, so a valid key without the db.download scope is refused as
// unauthorized, not as forbidden.
func explain(err error) error {
	var apiErr *internetdata.Error
	if !errors.As(err, &apiErr) {
		return err
	}
	switch apiErr.Kind {
	case internetdata.KindUnauthorized:
		return fmt.Errorf("%w\nthe API key was not accepted, or it lacks the db.download scope; check `%s whoami`",
			err, progBase)
	case internetdata.KindForbidden:
		return fmt.Errorf("%w\nthis organization's license does not cover that; `%s db list` shows what it holds",
			err, progBase)
	case internetdata.KindRateLimited:
		return fmt.Errorf("%w\nslow down, and try again shortly", err)
	}
	return err
}
