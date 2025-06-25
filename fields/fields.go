package fields

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

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

// builderStruct provides a fluent builder pattern for constructing
// field specifications for Tripletex API queries. It allows method chaining
// to build complex nested field structures.
//
// Example usage:
//
//	builder := Builder.New().
//		All().
//		Add("name").
//		Group("address", "street", "city").
//		Group("orders", Builder.New().Add("id").Add("total"))
//
//	result := builder.String() // "*,address(city,street),name,orders(id,total)"
type builderStruct struct {
	fields fields
}

// Builder is the global builder instance used to create new field builders.
// Use [builderStruct.New] to start building field specifications.
//
// Example:
//
//	fields := Builder.New().All().Add("name").String()
var Builder = &builderStruct{}

// New creates a new [Builder] instance with an empty field set.
// This is the starting point for building field specifications.
//
// Example:
//
//	builder := Builder.New()
//	fields := builder.Add("name").Add("email").String()
//	// Result: "email,name"
func (fb *builderStruct) New() *builderStruct {
	return &builderStruct{fields: make(fields)}
}

// All adds a wildcard field (*) to include all available fields.
// This is useful when you want to retrieve all fields for an entity.
//
// Example:
//
//	fields := Builder.New().All().String()
//	// Result: "*"
//
//	fields = Builder.New().All().Add("customField").String()
//	// Result: "*,customField"
func (fb *builderStruct) All() *builderStruct {
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
//	fields := Builder.New().Add("name").Add("email").String()
//	// Result: "email,name"
//
//	fields = Builder.New().Add("id").Add("firstName").Add("lastName").String()
//	// Result: "firstName,id,lastName"
func (fb *builderStruct) Add(name string) *builderStruct {
	if fb.fields == nil {
		fb.fields = make(fields)
	}
	fb.fields[name] = nil
	return fb
}

// Group creates a nested field group with the specified name and sub-fields.
// Sub-fields can be strings (simple fields), other [builderStruct] instances,
// or fields maps. This allows for complex nested field structures.
//
// Examples:
//
//	// Simple nested group
//	fields := Builder.New().Group("address", "street", "city").String()
//	// Result: "address(city,street)"
//
//	// Group with nested sub-groups
//	fields = Builder.New().
//		Group("customer", "name", "email",
//			Builder.New().Group("address", "street", "city")).String()
//	// Result: "customer(address(city,street),email,name)"
//
//	// Multiple groups with wildcard
//	fields = Builder.New().
//		All().
//		Group("orders", "*", Builder.New().Group("product", "id", "name")).
//		Group("contact", "firstName", "lastName").String()
//	// Result: "*,contact(firstName,lastName),orders(*,product(id,name))"
func (fb *builderStruct) Group(name string, f ...any) *builderStruct {
	if fb.fields == nil {
		fb.fields = make(fields)
	}

	nestedFields := make(fields)
	for _, field := range f {
		switch f := field.(type) {
		case string:
			nestedFields[f] = nil
		case *builderStruct:
			maps.Copy(nestedFields, f.fields)
		case fields:
			maps.Copy(nestedFields, f)
		}
	}

	fb.fields[name] = &nestedFields
	return fb
}

// String converts the [builderStruct] to a formatted field string
// suitable for use as a parameter in Tripletex API queries.
// This method implements the Stringer interface.
//
// Returns a comma-separated string with nested fields enclosed in parentheses.
// Fields are sorted alphabetically for consistent output.
//
// Examples:
//
//	// Simple fields
//	fields := Builder.New().Add("name").Add("email").String()
//	// Result: "email,name"
//
//	// With wildcard
//	fields = Builder.New().All().Add("customField").String()
//	// Result: "*,customField"
//
//	// Complex nested structure
//	fields = Builder.New().
//		All().
//		Group("customer", "name", "email",
//			Builder.New().Group("address", "street", "city")).
//		Group("orders", Builder.New().Add("id").Add("total")).String()
//	// Result: "*,customer(address(city,street),email,name),orders(id,total)"
func (fb *builderStruct) String() string {
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
