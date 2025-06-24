package tripletex

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

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

	t.Logf("Available companies with login access:")
	for i, company := range *res.JSON200.Values {
		t.Logf("  %d: ID=%d, Name=%s", i, *company.Id, *company.Name)
	}

	validClientId := *(*res.JSON200.Values)[0].Id
	t.Logf("Using clientId: %d", validClientId)
	clientId := int64(validClientId)
	c = New(creds, WithBaseURLOption(baseURL), WithAccountantClient(clientId))
	require.False(c.IsTokenValid(), "token should be invalid after init")
	require.NoError(c.CheckAuth(), "should be able to check token after init")
	require.True(c.IsTokenValid(), "token should be valid after check")

	// Test basic endpoint to verify auth works
	whoAmIRes2, err := c.TokenSessionWhoAmIWhoAmIWithResponse(context.Background(), &TokenSessionWhoAmIWhoAmIParams{})
	require.NoError(err, "whoAmI with accountant client should not error")
	require.NotNil(whoAmIRes2, "whoAmI with accountant client should not be nil")
	t.Logf("WhoAmI with accountant client status: %d", whoAmIRes2.StatusCode())
	if whoAmIRes2.StatusCode() != http.StatusOK {
		t.Logf("WhoAmI response body: %s", string(whoAmIRes2.Body))
	}

	departmentRes, err := c.DepartmentSearchWithResponse(context.Background(), &DepartmentSearchParams{})
	require.NoError(err, "departments search should not error")
	require.NotNil(departmentRes, "departments res should not be nil")

	if departmentRes.StatusCode() != http.StatusOK {
		t.Logf("Department search failed with status: %d", departmentRes.StatusCode())
		if departmentRes.Body != nil {
			t.Logf("Response body: %s", string(departmentRes.Body))
		}
	}

	require.Equal(http.StatusOK, departmentRes.StatusCode(), "departments status should be OK (200)")
	require.NotNil(departmentRes.JSON200, "departments res.JSON200 should not be nil")
	require.NotNil(departmentRes.JSON200.Values, "departments res.JSON200.Values should not be nil")
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
	require.NotNil(whoAmIRes.JSON200, "whoAmI JSON200 should not be nil")
	require.NotNil(whoAmIRes.JSON200.Value, "whoAmI value should not be nil")

	employeeId := whoAmIRes.JSON200.Value.EmployeeId
	actualEmployeeId := whoAmIRes.JSON200.Value.ActualEmployeeId

	require.NotNil(employeeId, "employeeId should not be nil")
	require.NotNil(actualEmployeeId, "actualEmployeeId should not be nil")

	t.Logf("Employee ID (owner): %d", *employeeId)
	t.Logf("Actual Employee ID (token): %d", *actualEmployeeId)

	// Step 2: Try to get entitlements - this may fail due to permissions
	ownerEntitlements, err := c.EmployeeEntitlementGetWithResponse(context.Background(), int64(*employeeId), &EmployeeEntitlementGetParams{})
	require.NoError(err, "owner entitlements should not error")
	t.Logf("Owner entitlements status: %d", ownerEntitlements.StatusCode())
	if ownerEntitlements.StatusCode() != http.StatusOK {
		t.Logf("Owner entitlements response body: %s", string(ownerEntitlements.Body))
		t.Skip("Cannot access entitlements - likely missing 'User admin' permission")
		return
	}

	tokenEntitlements, err := c.EmployeeEntitlementGetWithResponse(context.Background(), int64(*actualEmployeeId), &EmployeeEntitlementGetParams{})
	require.NoError(err, "token entitlements should not error")
	t.Logf("Token entitlements status: %d", tokenEntitlements.StatusCode())
	if tokenEntitlements.StatusCode() != http.StatusOK {
		t.Logf("Token entitlements response body: %s", string(tokenEntitlements.Body))
		t.Skip("Cannot access token entitlements - likely missing 'User admin' permission")
		return
	}

	require.NotNil(ownerEntitlements.JSON200, "owner entitlements JSON200 should not be nil")
	require.NotNil(ownerEntitlements.JSON200.Value, "owner entitlements value should not be nil")
	require.NotNil(tokenEntitlements.JSON200, "token entitlements JSON200 should not be nil")
	require.NotNil(tokenEntitlements.JSON200.Value, "token entitlements value should not be nil")

	// Check if both users have "User admin" entitlement
	ownerHasUserAdmin := ownerEntitlements.JSON200.Value.Name != nil && *ownerEntitlements.JSON200.Value.Name == "User admin"
	tokenHasUserAdmin := tokenEntitlements.JSON200.Value.Name != nil && *tokenEntitlements.JSON200.Value.Name == "User admin"

	t.Logf("Owner has 'User admin': %t", ownerHasUserAdmin)
	t.Logf("Token has 'User admin': %t", tokenHasUserAdmin)

	// Log entitlements for debugging
	if ownerEntitlements.JSON200.Value.Name != nil {
		t.Logf("Owner entitlement: %s", *ownerEntitlements.JSON200.Value.Name)
	}
	if tokenEntitlements.JSON200.Value.Name != nil {
		t.Logf("Token entitlement: %s", *tokenEntitlements.JSON200.Value.Name)
	}
}

func TestAccountantTokenEntitlements(t *testing.T) {
	require := require.New(t)

	baseURL := mustEnv("TRIPLETEX_BASE_URL_ACCOUNTANT")
	consumerToken := mustEnv("TRIPLETEX_CONSUMER_TOKEN_ACCOUNTANT")
	employeeToken := mustEnv("TRIPLETEX_EMPLOYEE_TOKEN_ACCOUNTANT")
	creds := Credentials{
		ConsumerToken: consumerToken,
		EmployeeToken: employeeToken,
	}

	c := New(creds, WithBaseURLOption(baseURL))
	require.NoError(c.CheckAuth())

	// Step 1: Get whoAmI to find employeeId and actualEmployeeId
	whoAmIRes, err := c.TokenSessionWhoAmIWhoAmIWithResponse(context.Background(), &TokenSessionWhoAmIWhoAmIParams{})
	require.NoError(err, "whoAmI should not error")
	require.NotNil(whoAmIRes.JSON200, "whoAmI JSON200 should not be nil")
	require.NotNil(whoAmIRes.JSON200.Value, "whoAmI value should not be nil")

	employeeId := whoAmIRes.JSON200.Value.EmployeeId
	actualEmployeeId := whoAmIRes.JSON200.Value.ActualEmployeeId

	require.NotNil(employeeId, "employeeId should not be nil")
	require.NotNil(actualEmployeeId, "actualEmployeeId should not be nil")

	t.Logf("Accountant Employee ID (owner): %d", *employeeId)
	t.Logf("Accountant Actual Employee ID (token): %d", *actualEmployeeId)

	// Step 2: Try to get entitlements - this may fail due to permissions
	ownerEntitlements, err := c.EmployeeEntitlementGetWithResponse(context.Background(), int64(*employeeId), &EmployeeEntitlementGetParams{})
	require.NoError(err, "owner entitlements should not error")
	t.Logf("Accountant Owner entitlements status: %d", ownerEntitlements.StatusCode())
	if ownerEntitlements.StatusCode() != http.StatusOK {
		t.Logf("Accountant Owner entitlements response body: %s", string(ownerEntitlements.Body))
		t.Skip("Cannot access accountant entitlements - likely missing 'User admin' permission")
		return
	}

	require.NotNil(ownerEntitlements.JSON200, "owner entitlements JSON200 should not be nil")
	require.NotNil(ownerEntitlements.JSON200.Value, "owner entitlements value should not be nil")

	// Check if user has "User admin" entitlement
	ownerHasUserAdmin := ownerEntitlements.JSON200.Value.Name != nil && *ownerEntitlements.JSON200.Value.Name == "User admin"
	t.Logf("Accountant Owner has 'User admin': %t", ownerHasUserAdmin)

	// Log entitlements for debugging
	if ownerEntitlements.JSON200.Value.Name != nil {
		t.Logf("Accountant Owner entitlement: %s", *ownerEntitlements.JSON200.Value.Name)
	}
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
		input       *fieldsBuilderStruct
		expected    string
	}{
		{
			description: "short",
			input:       FieldsBuilder.New().All().Group("orders", "id", FieldsBuilder.New().Group("project", "id")),
			expected:    "*,orders(id,project(id))",
		},
		{
			description: "long (OrderSearch)",
			input: FieldsBuilder.New().All().
				Group("contact", "id", "firstName", "lastName").
				Group("customer", "id").
				Group("deliveryAddress", "*", "country").
				Group("department", "id").
				Group("preliminaryInvoice", "*").
				Group("ourContact", "id", "firstName", "lastName").
				Group("orderLines", "*", FieldsBuilder.New().Group("product", "number")).
				Group("project", "id"),
			expected: "*,contact(firstName,id,lastName),customer(id),deliveryAddress(*,country),department(id),orderLines(*,product(number)),ourContact(firstName,id,lastName),preliminaryInvoice(*),project(id)",
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			require := require.New(t)
			result := tt.input.String()
			require.Equal(tt.expected, result)
		})
	}
}
