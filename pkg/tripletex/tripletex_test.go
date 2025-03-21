package tripletex

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

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
}

// Require enviornment variable. Panics if not found.
func mustEnv(env string) string {
	v, ok := os.LookupEnv(env)
	if !ok {
		panic(fmt.Sprintf("%s is not set", env))
	}

	return v
}
