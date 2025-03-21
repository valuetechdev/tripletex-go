//go:generate go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config cfg.yaml ../../api/openapi.json

package tripletex

import (
	"fmt"
	"net/http"
)

type APIClient = ClientWithResponses

type APIClientOpts struct {
	BaseUrl    string // Defaults to "https://tripletex.tech/v2"
	HttpClient *http.Client
}

// Returns APIClient for Tripletex.
//
// You can provide a different baseUrl if you want to work against Tripletex's
// test environement via opts.
//
// You can reuse an already generated token and have it revalidated if it has
// expired, by providing it via opts.
//
// You can provide a custom http.Client via opts.
//
// Returns error if fails to initialize client.
func New(token *Token, opts *APIClientOpts) (*APIClient, error) {
	if opts.BaseUrl == "" {
		opts.BaseUrl = "https://tripletex.no/v2"
	}
	httpClient := http.DefaultClient
	if opts.HttpClient != nil {
		httpClient = opts.HttpClient
	}

	c, err := NewClientWithResponses(
		opts.BaseUrl,
		WithRequestEditorFn(token.InterceptAuth),
		WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("tripletex: failed to create new client: %w", err)
	}

	return &APIClient{c}, nil
}
