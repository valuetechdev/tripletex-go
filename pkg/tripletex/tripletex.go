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
	httpClient    *http.Client
	*ClientWithResponses
}

type Credentials struct {
	BaseURL       string // Defaults to "https://tripletex.tech/v2"
	ConsumerToken string // Application specific token
	EmployeeToken string // Client specific token
}

type Opts struct {
	HttpClient    *http.Client  // Defaults to [http.DefaultClient]
	TokenDuration time.Duration // Defaults to one month
}

// Returns new TripletexClient.
//
// You can provide a different BaseURL if you want to work against Tripletex's
// test environement via opts.
//
// You can reuse an already generated token and have it revalidated if it has
// expired, by using [TripletexClient.SetToken()].
//
// You can provide a custom http.Client via opts.
func New(credentials Credentials, opts *Opts) *TripletexClient {
	now := time.Now()
	client := &TripletexClient{
		credentials:   credentials,
		tokenDuration: now.AddDate(0, 1, 0).Sub(now),
		httpClient:    http.DefaultClient,
	}
	if client.credentials.BaseURL == "" {
		client.credentials.BaseURL = "https://tripletex.no/v2"
	}
	if opts != nil {
		if dur := opts.TokenDuration; dur != 0 {
			client.tokenDuration = dur
		}
		if opts.HttpClient != nil {
			client.httpClient = opts.HttpClient
		}
	}

	c, err := NewClientWithResponses(
		client.credentials.BaseURL,
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
