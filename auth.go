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
	ExpiresAt   time.Time `json:"expiresAt"`
	AccessToken string    `json:"token"`
}

// Revalidates Token.
//
// Returns error when failing to make http requests, read/parse response body.
func (c *TripletexClient) revalidate() error {
	creds := c.credentials
	expiresAt := time.Now().Add(c.tokenDuration)
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/token/session/:create", c.baseURL), http.NoBody)
	if err != nil {
		return fmt.Errorf("tripletex: auth: failed to create http request: %w", err)
	}

	q := req.URL.Query()
	q.Add("consumerToken", creds.ConsumerToken)
	q.Add("employeeToken", creds.EmployeeToken)
	q.Add("expirationDate", expiresAt.Format(time.DateOnly))
	req.URL.RawQuery = q.Encode()

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("tripletex: auth: failed to do http request: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("tripletex: auth: status not OK: %s", res.Status)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("tripletex: auth: failed to read response body: %w", err)
	}

	var sessionTokenRes ResponseWrapperSessionToken
	if err = json.Unmarshal(body, &sessionTokenRes); err != nil {
		return fmt.Errorf("tripletex: auth: failed to parse response body: %w", err)
	}

	if sessionTokenRes.Value == nil {
		return fmt.Errorf("tripletex: auth: session token value body is empty")
	}
	sessionToken := *sessionTokenRes.Value

	if sessionTokenRes.Value.ExpirationDate == nil {
		return fmt.Errorf("tripletex: auth: session token expirationDate is empty")
	}

	expiresAt, err = time.Parse(time.DateOnly, *sessionToken.ExpirationDate)
	if err != nil {
		return fmt.Errorf("authentication: failed to parse expiresAt (%s): %w", *sessionToken.ExpirationDate, err)
	}

	c.token = &Token{
		AccessToken: *sessionToken.Token,
		ExpiresAt:   expiresAt,
	}

	return nil
}

func (c *TripletexClient) GetToken() *Token {
	return c.token
}

func (c *TripletexClient) SetToken(token *Token) {
	c.token = token
}

// Returns true if token is valid.
func (c *TripletexClient) IsTokenValid() bool {
	if c.token == nil {
		return false
	}
	return time.Now().Before(c.token.ExpiresAt)
}

// Check if auth is valid.
//
// Revalidates token if invalid
func (c *TripletexClient) CheckAuth() error {
	if !c.IsTokenValid() {
		if err := c.revalidate(); err != nil {
			return fmt.Errorf("tripletex: auth: failed to revalidate token: %w", err)
		}
	}
	return nil
}

// Intercepts authentication on http request r.
//
// Sets the token as basic auth with username 0 and token.AccessToken as password.
//
// Returns error if unable to revalidate token.
func (c *TripletexClient) interceptAuth(ctx context.Context, r *http.Request) error {
	if err := c.CheckAuth(); err != nil {
		return err
	}
	username := "0"
	r.SetBasicAuth(username, c.token.AccessToken)

	return nil
}
