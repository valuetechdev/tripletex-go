package tripletex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/valuetechdev/tripletex-go/fields"
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
	f := fields.Builder.Add("name").Add("id").String()
	customersRes, err := c.CustomerSearchWithResponse(context.Background(), &CustomerSearchParams{ChangedSince: &lastYearString, Fields: &f})
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
	require.NotNil(res.JSONDefault, "companies JSONDefault with login should not be nil")
	require.NotNil(res.JSONDefault.Values, "companies JSONDefault.Values with login should not be nil")

	whoAmIRes, err := c.TokenSessionWhoAmIWhoAmIWithResponse(context.Background(), &TokenSessionWhoAmIWhoAmIParams{})
	require.NoError(err, "should not error for checking whoAmI")
	require.NotNil(whoAmIRes, "whoAmI should not be nil")
	require.Equal(http.StatusOK, whoAmIRes.StatusCode(), "whoAmI status should be OK (200)")
	require.NotNil(whoAmIRes.JSONDefault, "whoAmI.JSONDefault value should not be nil")
	require.NotNil(whoAmIRes.JSONDefault.Value, "whoAmI.JSONDefault.Value should not be nil")

	t.Logf("Available companies with login access:")
	for i, company := range *res.JSONDefault.Values {
		t.Logf("  %d: ID=%d, Name=%s", i, *company.Id, *company.Name)
	}
	validClientId := *(*res.JSONDefault.Values)[0].Id
	t.Logf("Using clientId: %d", validClientId)
	clientId := int64(validClientId)
	c = New(creds, WithBaseURLOption(baseURL), WithAccountantClient(clientId))
	require.False(c.IsTokenValid(), "token should be invalid after init")
	require.NoError(c.CheckAuth(), "should be able to check token after init")
	require.True(c.IsTokenValid(), "token should be valid after check")

	whoAmIRes2, err := c.TokenSessionWhoAmIWhoAmIWithResponse(context.Background(), &TokenSessionWhoAmIWhoAmIParams{})
	require.NoError(err, "whoAmI with accountant client should not error")
	require.NotNil(whoAmIRes2, "whoAmI with accountant client should not be nil")
	require.Equal(http.StatusOK, whoAmIRes2.StatusCode(), "whoAmI status should be OK (200)")

	departmentRes, err := c.DepartmentSearchWithResponse(context.Background(), &DepartmentSearchParams{})
	require.NoError(err, "departments search should not error")
	require.NotNil(departmentRes, "departments res should not be nil")

	if departmentRes.StatusCode() != http.StatusOK {
		t.Logf("departments search failed with status: %d", departmentRes.StatusCode())
		if departmentRes.Body != nil {
			t.Logf("Response body: %s", string(departmentRes.Body))
		}
	}

	require.Equal(http.StatusOK, departmentRes.StatusCode(), "departments status should be OK (200)")
	require.NotNil(departmentRes.JSONDefault, "departments res.JSONDefault should not be nil")
	require.NotNil(departmentRes.JSONDefault.Values, "departments res.JSONDefault.Values should not be nil")
}

func TestEmployeeTokenEntitlements(t *testing.T) {
	require := require.New(t)

	baseURL := mustEnv("TRIPLETEX_BASE_URL")
	consumerToken := mustEnv("TRIPLETEX_CONSUMER_TOKEN")
	employeeToken := mustEnv("TRIPLETEX_EMPLOYEE_TOKEN")
	creds := Credentials{
		ConsumerToken: consumerToken,
		EmployeeToken: employeeToken,
	}

	c := New(creds, WithBaseURLOption(baseURL))
	require.NoError(c.CheckAuth())

	// Step 1: Get whoAmI to find employeeId and actualEmployeeId
	whoAmIRes, err := c.TokenSessionWhoAmIWhoAmIWithResponse(context.Background(), &TokenSessionWhoAmIWhoAmIParams{})
	require.NoError(err, "whoAmI should not error")
	require.NotNil(whoAmIRes.JSONDefault, "whoAmI JSONDefault should not be nil")
	require.NotNil(whoAmIRes.JSONDefault.Value, "whoAmI value should not be nil")
	j, _ := json.MarshalIndent(whoAmIRes.JSONDefault.Value, "", "  ")
	t.Logf("whoAmIRes: %s", string(j))

	employeeId := int64(*whoAmIRes.JSONDefault.Value.EmployeeId)
	actualEmployeeId := int64(*whoAmIRes.JSONDefault.Value.ActualEmployeeId)

	require.NotNil(employeeId, "employeeId should not be nil")
	require.NotNil(actualEmployeeId, "actualEmployeeId should not be nil")

	t.Logf("Employee ID (owner): %d", employeeId)
	t.Logf("Actual Employee ID (token): %d", actualEmployeeId)

	// Step 2: Try to get entitlements - this may fail due to permissions
	ownerEntitlements, err := c.EmployeeEntitlementSearchWithResponse(context.Background(), &EmployeeEntitlementSearchParams{EmployeeId: &employeeId})
	require.NoError(err, "owner entitlements should not error")
	require.Equal(http.StatusOK, ownerEntitlements.StatusCode(), "owner entitlements status should be OK (200)")

	tokenEntitlements, err := c.EmployeeEntitlementSearchWithResponse(context.Background(), &EmployeeEntitlementSearchParams{EmployeeId: &actualEmployeeId})
	require.NoError(err, "token entitlements should not error")
	require.Equal(http.StatusOK, tokenEntitlements.StatusCode(), "token entitlements status should be OK (200)")

	require.NotNil(ownerEntitlements.JSONDefault, "owner entitlements JSONDefault should not be nil")
	require.NotNil(ownerEntitlements.JSONDefault.Values, "owner entitlements value should not be nil")
	require.NotNil(tokenEntitlements.JSONDefault, "token entitlements JSONDefault should not be nil")
	require.NotNil(tokenEntitlements.JSONDefault.Values, "token entitlements value should not be nil")

	// Check if both users have "User admin" entitlement
	hasUserAdmin := func(entitlements *[]Entitlement) bool {
		for _, entitlement := range *entitlements {
			if entitlement.Name != nil && *entitlement.Name == "User admin" {
				return true
			}
		}
		return false
	}

	ownerHasUserAdmin := hasUserAdmin(ownerEntitlements.JSONDefault.Values)
	tokenHasUserAdmin := hasUserAdmin(tokenEntitlements.JSONDefault.Values)

	t.Logf("Owner has 'User admin': %t", ownerHasUserAdmin)
	t.Logf("Token has 'User admin': %t", tokenHasUserAdmin)

	// Log all entitlements for debugging
	t.Log("Owner entitlements:")
	for _, entitlement := range *ownerEntitlements.JSONDefault.Values {
		if entitlement.Name != nil {
			t.Logf("  - %s", *entitlement.Name)
		}
	}

	t.Log("Token entitlements:")
	for _, entitlement := range *tokenEntitlements.JSONDefault.Values {
		if entitlement.Name != nil {
			t.Logf("  - %s", *entitlement.Name)
		}
	}
}

func TestBadResponse(t *testing.T) {
	require := require.New(t)

	baseURL := mustEnv("TRIPLETEX_BASE_URL")
	consumerToken := mustEnv("TRIPLETEX_CONSUMER_TOKEN")
	employeeToken := mustEnv("TRIPLETEX_EMPLOYEE_TOKEN")
	creds := Credentials{
		ConsumerToken: consumerToken,
		EmployeeToken: employeeToken,
	}

	c := New(creds, WithBaseURLOption(baseURL))
	require.False(c.IsTokenValid(), "token should be invalid after init")
	require.NoError(c.CheckAuth(), "should be able to check token after init")
	require.True(c.IsTokenValid(), "token should be valid after check")

	sorting := "yolo"
	from := -69
	res, err := c.CompanyWithLoginAccessGetWithLoginAccessWithResponse(context.Background(), &CompanyWithLoginAccessGetWithLoginAccessParams{Sorting: &sorting, From: &from})
	require.NoError(err, "should not error for getting companies with login")
	require.NotNil(res, "companies with login should not be nil")
	require.Equal(&ListResponseClient{}, res.JSONDefault, "zero value should be equal to res.JSONDefault")
	require.Equal(http.StatusUnprocessableEntity, res.StatusCode(), "status code should be 422")
	require.Nil(res.JSONDefault.Values, "JSONDefault.Values should be nil")
}

// Require environment variable. Panics if not found.
func mustEnv(env string) string {
	v, ok := os.LookupEnv(env)
	if !ok {
		panic(fmt.Sprintf("%s is not set", env))
	}

	return v
}
