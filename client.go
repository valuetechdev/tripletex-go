//go:generate go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config cfg.yaml ./api/openapi.json

package tripletex

import (
	"fmt"
	"maps"
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
	clientId      int64  // Used with [WithAccountantClient]
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

// WithBaseURLOption sets a custom base URL. Defaults to "https://tripletex.no/v2".
func WithBaseURLOption(baseURL string) Option {
	return func(tc *TripletexClient) {
		tc.baseURL = baseURL
	}
}

// WithAccountantClient sets clientId as username for
// [TripletexClient.interceptAuth].
//
// See https://developer.tripletex.no/docs/documentation/authentication-and-tokens/#accountant-token
// for more details.
func WithAccountantClient(clientId int64) Option {
	return func(tc *TripletexClient) {
		tc.credentials.clientId = clientId
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

// fields represents a nested field structure for Tripletex API queries.
// It maps field names to nested field structures, where nil values indicate
// simple fields without nesting.
//
// Example:
//
//	fields{
//		"*": nil,           // Include all fields
//		"name": nil,        // Simple field
//		"address": &fields{ // Nested field group
//			"street": nil,
//			"city": nil,
//		},
//	}
type fields map[string]*fields

// fieldsBuilderStruct provides a fluent builder pattern for constructing
// field specifications for Tripletex API queries. It allows method chaining
// to build complex nested field structures.
//
// Example usage:
//
//	builder := FieldsBuilder.New().
//		All().
//		Add("name").
//		Group("address", "street", "city").
//		Group("orders", FieldsBuilder.New().Add("id").Add("total"))
//
//	result := builder.String() // "*,address(city,street),name,orders(id,total)"
type fieldsBuilderStruct struct {
	fields fields
}

// FieldsBuilder is the global builder instance used to create new field builders.
// Use FieldsBuilder.New() to start building field specifications.
//
// Example:
//
//	fields := FieldsBuilder.New().All().Add("name").String()
var FieldsBuilder = &fieldsBuilderStruct{}

// New creates a new FieldsBuilder instance with an empty field set.
// This is the starting point for building field specifications.
//
// Example:
//
//	builder := FieldsBuilder.New()
//	fields := builder.Add("name").Add("email").String()
//	// Result: "email,name"
func (fb *fieldsBuilderStruct) New() *fieldsBuilderStruct {
	return &fieldsBuilderStruct{fields: make(fields)}
}

// All adds a wildcard field (*) to include all available fields.
// This is useful when you want to retrieve all fields for an entity.
//
// Example:
//
//	fields := FieldsBuilder.New().All().String()
//	// Result: "*"
//
//	fields = FieldsBuilder.New().All().Add("customField").String()
//	// Result: "*,customField"
func (fb *fieldsBuilderStruct) All() *fieldsBuilderStruct {
	if fb.fields == nil {
		fb.fields = make(fields)
	}
	fb.fields["*"] = nil
	return fb
}

// Add adds a simple field to the field specification.
// Simple fields are those without nested sub-fields.
//
// Example:
//
//	fields := FieldsBuilder.New().Add("name").Add("email").String()
//	// Result: "email,name"
//
//	fields = FieldsBuilder.New().Add("id").Add("firstName").Add("lastName").String()
//	// Result: "firstName,id,lastName"
func (fb *fieldsBuilderStruct) Add(name string) *fieldsBuilderStruct {
	if fb.fields == nil {
		fb.fields = make(fields)
	}
	fb.fields[name] = nil
	return fb
}

// Group creates a nested field group with the specified name and sub-fields.
// Sub-fields can be strings (simple fields), other fieldsBuilderStruct instances,
// or fields maps. This allows for complex nested field structures.
//
// Examples:
//
//	// Simple nested group
//	fields := FieldsBuilder.New().Group("address", "street", "city").String()
//	// Result: "address(city,street)"
//
//	// Group with nested sub-groups
//	fields = FieldsBuilder.New().
//		Group("customer", "name", "email",
//			FieldsBuilder.New().Group("address", "street", "city")).String()
//	// Result: "customer(address(city,street),email,name)"
//
//	// Multiple groups with wildcard
//	fields = FieldsBuilder.New().
//		All().
//		Group("orders", "*", FieldsBuilder.New().Group("product", "id", "name")).
//		Group("contact", "firstName", "lastName").String()
//	// Result: "*,contact(firstName,lastName),orders(*,product(id,name))"
func (fb *fieldsBuilderStruct) Group(name string, f ...any) *fieldsBuilderStruct {
	if fb.fields == nil {
		fb.fields = make(fields)
	}

	nestedFields := make(fields)
	for _, field := range f {
		switch f := field.(type) {
		case string:
			nestedFields[f] = nil
		case *fieldsBuilderStruct:
			maps.Copy(nestedFields, f.fields)
		case fields:
			maps.Copy(nestedFields, f)
		}
	}

	fb.fields[name] = &nestedFields
	return fb
}

// String converts the fieldsBuilderStruct to a formatted field string
// suitable for use as a parameter in Tripletex API queries.
// This method implements the Stringer interface.
//
// Returns a comma-separated string with nested fields enclosed in parentheses.
// Fields are sorted alphabetically for consistent output.
//
// Examples:
//
//	// Simple fields
//	fields := FieldsBuilder.New().Add("name").Add("email").String()
//	// Result: "email,name"
//
//	// With wildcard
//	fields = FieldsBuilder.New().All().Add("customField").String()
//	// Result: "*,customField"
//
//	// Complex nested structure
//	fields = FieldsBuilder.New().
//		All().
//		Group("customer", "name", "email",
//			FieldsBuilder.New().Group("address", "street", "city")).
//		Group("orders", FieldsBuilder.New().Add("id").Add("total")).String()
//	// Result: "*,customer(address(city,street),email,name),orders(id,total)"
func (fb *fieldsBuilderStruct) String() string {
	return fieldsToString(fb.fields)
}

// fieldsToString converts a fields map to a formatted string representation.
// This is a helper function used internally by the String() method.
// It recursively processes nested field structures and returns a comma-separated
// string with nested fields enclosed in parentheses, sorted alphabetically.
func fieldsToString(input fields) string {
	var s []string
	for k, v := range input {
		if v != nil {
			s = append(s, fmt.Sprintf("%s(%s)", k, fieldsToString(*v)))
		} else {
			s = append(s, k)
		}
	}

	slices.Sort(s)
	return strings.Join(s, ",")
}
