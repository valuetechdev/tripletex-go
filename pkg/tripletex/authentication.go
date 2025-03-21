package tripletex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Token struct {
	ExpiresAt time.Time  `json:"expiresAt"`
	Token     string     `json:"token"`
	opts      *TokenOpts `json:"-"`
}

type TokenOpts struct {
	BaseUrl       string     // Defaults to "https://tripletex.tech/v2"
	ConsumerToken string     // Application specific token
	EmployeeToken string     // Client specific token
	ExpiresAt     *time.Time // Uses only time.DateOnly
}

// Returns new Token. Is expected to be used with InterceptAuth() method.
func NewToken(opts *TokenOpts) *Token {
	return &Token{
		opts: opts,
	}
}

// Revalidates Token.
//
// Returns error when failing to make http requests, read/parse response body.
func (t *Token) revalidate() error {
	if t.opts.BaseUrl == "" {
		t.opts.BaseUrl = "https://tripletex.no/v2"
	}

	if t.opts.ExpiresAt == nil {
		oneMonthAhead := time.Now().AddDate(0, 1, 0)
		t.opts.ExpiresAt = &oneMonthAhead
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/token/session/:create", t.opts.BaseUrl), http.NoBody)
	if err != nil {
		return fmt.Errorf("authentication: failed to create http request: %w", err)
	}

	q := req.URL.Query()
	q.Add("consumerToken", t.opts.ConsumerToken)
	q.Add("employeeToken", t.opts.EmployeeToken)
	q.Add("expirationDate", t.opts.ExpiresAt.Format(time.DateOnly))
	req.URL.RawQuery = q.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("authentication: failed to do http request: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication: status not OK: %s: %w", res.Status, err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("authentication: failed to read response body: %w", err)
	}

	var sessionTokenRes ResponseWrapperSessionToken
	if err = json.Unmarshal(body, &sessionTokenRes); err != nil {
		return fmt.Errorf("authentication: failed to parse response body: %w", err)
	}

	if sessionTokenRes.Value == nil {
		return fmt.Errorf("authentication: session token value body is empty")
	}
	sessionToken := sessionTokenRes.Value

	expiresAt, err := time.Parse(time.DateOnly, sessionToken.ExpirationDate)
	if err != nil {
		return fmt.Errorf("authentication: failed to parse expiresAt (%s): %w", sessionToken.ExpirationDate, err)
	}

	t.Token = sessionToken.Token
	t.ExpiresAt = expiresAt

	return nil
}

// Returns true if t has expired.
func (t *Token) HasExpired() bool {
	return time.Now().UTC().After(t.ExpiresAt)
}

// Check if auth is valid
func (t *Token) CheckAuth() error {
	if t.HasExpired() {
		if err := t.revalidate(); err != nil {
			return fmt.Errorf("authentication: failed to revalidate token: %w", err)
		}
	}
	return nil
}

// Intercepts authentication on http request r.
//
// Sets the token as basic auth with username 0 and t.Token as password.
//
// Returns error if unable to revalidate token.
func (t *Token) InterceptAuth(ctx context.Context, r *http.Request) error {
	if err := t.CheckAuth(); err != nil {
		return err
	}
	username := "0"
	r.SetBasicAuth(username, t.Token)

	return nil
}
