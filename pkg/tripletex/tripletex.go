//go:generate go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config cfg.yaml ../../api/openapi.json

package tripletex

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"
)

type TripletexClient struct {
	token         *Token
	tokenDuration time.Duration
	credentials   Credentials
	baseURL       string
	httpClient    *http.Client
	*ClientWithResponses
}

type Credentials struct {
	ConsumerToken string // Application specific token
	EmployeeToken string // Client specific token
}

type Option func(*TripletexClient)

// WithHttpClient sets a custom http.Client. Defaults to [http.DefaultClient].
func WithHttpClient(client *http.Client) Option {
	return func(tc *TripletexClient) {
		tc.httpClient = client
	}
}

// WithTokenDuration sets the token duration. Defaults to one month.
func WithTokenDuration(duration time.Duration) Option {
	return func(tc *TripletexClient) {
		tc.tokenDuration = duration
	}
}

// WithBaseURL sets a custom http.Client. Defaults to [http.DefaultClient].
func WithBaseURLOption(baseURL string) Option {
	return func(tc *TripletexClient) {
		tc.baseURL = baseURL
	}
}

// Returns new [TripletexClient].
//
// You can reuse an already generated token and have it revalidated if it has
// expired, by using [TripletexClient.SetToken].
//
// You can provide options to customize the client behavior.
func New(credentials Credentials, options ...Option) *TripletexClient {
	now := time.Now()
	client := &TripletexClient{
		baseURL:       "https://tripletex.no/v2",
		credentials:   credentials,
		tokenDuration: now.AddDate(0, 1, 0).Sub(now),
		httpClient:    http.DefaultClient,
	}

	for _, option := range options {
		option(client)
	}

	c, err := NewClientWithResponses(
		client.baseURL,
		WithRequestEditorFn(client.interceptAuth),
		WithHTTPClient(client.httpClient))
	if err != nil {
		panic(fmt.Errorf("tripletex: failed to create new client: %w", err))
	}

	client.ClientWithResponses = c
	return client
}

// Used with FieldsBuilder
type Fields map[string]*Fields

// Builds field string to use as parameter for queries to Tripletex.
//
// Returns a valid string.
//
// Example:
//
//	f := Fields{
//		"*": nil,
//		"orders": &Fields{
//			"id": nil,
//			"project": &Fields{
//				"id": nil,
//			},
//		},
//	}
//	fmt.Println(FieldsBuilder(f))
//	// Output: *,orders(id,project(id))
func FieldsBuilder(input Fields) string {
	var s []string
	for k, v := range input {
		if v != nil {
			s = append(s, fmt.Sprintf("%s(%s)", k, FieldsBuilder(*v)))
		} else {
			s = append(s, k)
		}
	}

	slices.Sort(s)
	return strings.Join(s, ",")
}
