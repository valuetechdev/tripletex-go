package fieldsbuilder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFieldsBuilder(t *testing.T) {
	for _, tt := range []struct {
		description string
		input       *builderStruct
		expected    string
	}{
		{
			description: "with single",
			input:       New().Single("id"),
			expected:    "id",
		},
		{
			description: "with two groups",
			input:       New().All().Group("orders", "id", New().Group("project", "id")),
			expected:    "*,orders(id,project(id))",
		},
		{
			description: "with multiple groups (OrderSearch)",
			input: New().All().
				Group("contact", "id", "firstName", "lastName").
				Group("customer", "id").
				Group("deliveryAddress", "*", "country").
				Group("department", "id").
				Group("preliminaryInvoice", "*").
				Group("ourContact", "id", "firstName", "lastName").
				Group("orderLines", "*", New().Group("product", "number")).
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
