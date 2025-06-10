package tripletex

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	require := require.New(t)

	baseUrl := mustEnv("TRIPLETEX_BASE_URL")
	consumerToken := mustEnv("TRIPLETEX_CONSUMER_TOKEN")
	employeeToken := mustEnv("TRIPLETEX_EMPLOYEE_TOKEN")
	token := NewToken(&TokenOpts{
		BaseUrl:       baseUrl,
		ConsumerToken: consumerToken,
		EmployeeToken: employeeToken,
	})

	c, err := New(token, &APIClientOpts{BaseUrl: baseUrl})
	require.NoError(err)

	lastYear, err := time.Parse(time.DateOnly, "2024-01-01")
	require.NoError(err)
	lastYearString := lastYear.Format(time.RFC3339)
	customersRes, err := c.CustomerSearchWithResponse(context.Background(), &CustomerSearchParams{ChangedSince: &lastYearString})
	require.NoError(err)
	require.NotNil(customersRes)

	require.NoError(token.CheckAuth())
}

func TestTokenAuth(t *testing.T) {
	require := require.New(t)

	baseUrl := mustEnv("TRIPLETEX_BASE_URL")
	consumerToken := mustEnv("TRIPLETEX_CONSUMER_TOKEN")
	employeeToken := mustEnv("TRIPLETEX_EMPLOYEE_TOKEN")
	token := NewToken(&TokenOpts{
		BaseUrl:       baseUrl,
		ConsumerToken: consumerToken,
		EmployeeToken: employeeToken,
	})

	require.NoError(token.CheckAuth())

	tmp, err := time.Parse(time.DateOnly, "2025-01-01")
	require.NoError(err)
	token.ExpiresAt = tmp
	require.NoError(token.CheckAuth())
}

// Require enviornment variable. Panics if not found.
func mustEnv(env string) string {
	v, ok := os.LookupEnv(env)
	if !ok {
		panic(fmt.Sprintf("%s is not set", env))
	}

	return v
}

func TestFieldsBuilder(t *testing.T) {
	for _, tt := range []struct {
		description string
		input       Fields
		expected    string
	}{
		{
			description: "short",
			input: Fields{
				"*": nil,
				"orders": &Fields{
					"id": nil,
					"project": &Fields{
						"id": nil,
					},
				},
			},
			expected: "*,orders(id,project(id))",
		},
		{
			description: "long (OrderSearch)",
			input: Fields{
				"*": nil,
				"contact": &Fields{
					"id":        nil,
					"firstName": nil,
					"lastName":  nil,
				},
				"customer": &Fields{"id": nil},
				"deliveryAddress": &Fields{
					"*":       nil,
					"country": nil,
				},
				"department":         &Fields{"id": nil},
				"preliminaryInvoice": &Fields{"*": nil},
				"ourContact": &Fields{
					"id":        nil,
					"firstName": nil,
					"lastName":  nil,
				},
				"orderLines": &Fields{
					"*": nil,
					"product": &Fields{
						"number": nil,
					},
				},
				"project": &Fields{"id": nil},
			},
			expected: "*,contact(firstName,id,lastName),customer(id),deliveryAddress(*,country),department(id),orderLines(*,product(number)),ourContact(firstName,id,lastName),preliminaryInvoice(*),project(id)",
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			assert := assert.New(t)
			result := FieldsBuilder(tt.input)
			assert.Equal(tt.expected, result)
		})
	}
}
