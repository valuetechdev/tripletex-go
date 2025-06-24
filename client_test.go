package tripletex

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	require := require.New(t)

	baseURL := mustEnv("TRIPLETEX_BASE_URL")
	consumerToken := mustEnv("TRIPLETEX_CONSUMER_TOKEN")
	employeeToken := mustEnv("TRIPLETEX_EMPLOYEE_TOKEN")
	creds := Credentials{
		ConsumerToken: consumerToken,
		EmployeeToken: employeeToken,
	}

	c := New(creds, WithBaseURLOption(baseURL))

	require.False(c.IsTokenValid())
	require.NoError(c.CheckAuth())

	lastYear, err := time.Parse(time.DateOnly, "2024-01-01")
	require.NoError(err)
	lastYearString := lastYear.Format(time.RFC3339)
	customersRes, err := c.CustomerSearchWithResponse(context.Background(), &CustomerSearchParams{ChangedSince: &lastYearString})
	require.NoError(err)
	require.NotNil(customersRes)
}

func TestNewClientWithAccountantClient(t *testing.T) {
	require := require.New(t)

	baseURL := mustEnv("TRIPLETEX_BASE_URL_ACCOUNTANT")
	consumerToken := mustEnv("TRIPLETEX_CONSUMER_TOKEN_ACCOUNTANT")
	employeeToken := mustEnv("TRIPLETEX_EMPLOYEE_TOKEN_ACCOUNTANT")
	creds := Credentials{
		ConsumerToken: consumerToken,
		EmployeeToken: employeeToken,
	}

	c := New(creds, WithBaseURLOption(baseURL))
	require.False(c.IsTokenValid(), "token should be invalid after init")
	require.NoError(c.CheckAuth(), "should be able to check token after init")
	require.True(c.IsTokenValid(), "token should be valid after check")

	res, err := c.CompanyWithLoginAccessGetWithLoginAccessWithResponse(context.Background(), &CompanyWithLoginAccessGetWithLoginAccessParams{})
	require.NoError(err, "should not error for getting companies with login")
	require.NotNil(res, "companies with login should not be nil")
	require.NotNil(res.JSON200, "companies JSON200 with login should not be nil")
	require.NotNil(res.JSON200.Values, "companies JSON200.Values with login should not be nil")

	whoAmIRes, err := c.TokenSessionWhoAmIWhoAmIWithResponse(context.Background(), &TokenSessionWhoAmIWhoAmIParams{})
	require.NoError(err, "should not error for checking whoAmI")
	require.NotNil(whoAmIRes, "whoAmI should not be nil")
	require.NotNil(whoAmIRes.JSON200, "whoAmI.JSON200 value should not be nil")
	require.NotNil(whoAmIRes.JSON200.Value, "whoAmI.JSON200.Value should not be nil")

	validClientId := *(*res.JSON200.Values)[0].Id
	t.Logf("validClientId: %d", validClientId)
	clientId := int64(validClientId)
	c = New(creds, WithBaseURLOption(baseURL), WithAccountantClient(clientId))
	require.False(c.IsTokenValid(), "token should be invalid after init")
	require.NoError(c.CheckAuth(), "should be able to check token after init")
	require.True(c.IsTokenValid(), "token should be valid after check")

	departmentRes, err := c.DepartmentSearchWithResponse(context.Background(), &DepartmentSearchParams{})
	require.NoError(err, "departments search should not error")
	require.NotNil(departmentRes, "departments res should not be nil")
	require.Equal(http.StatusOK, departmentRes.StatusCode(), "departments status should be OK (200)")
	require.NotNil(departmentRes.JSON200, "departments res.JSON200 should not be nil")
	require.NotNil(departmentRes.JSON200.Values, "departments res.JSON200.Values should not be nil")
}

// Require environment variable. Panics if not found.
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
