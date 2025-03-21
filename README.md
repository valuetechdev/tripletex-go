[![go reference badge](https://pkg.go.dev/badge/github.com/valuetechdev/tripletex-go.svg)](https://pkg.go.dev/github.com/valuetechdev/tripletex-go)

# tripletex-go

Go API client for [Tripletex]. It's generated with [oapi-codegen]

## Prerequisites

1. `consumerToken`
2. `employeeToken`

## Usage

```bash
go get github.com/valuetechdev/tripletex-go/tripletex
```

```go
import "github.com/valuetechdev/tripletex-go/tripletex"

func Whatever() error {
	token := NewToken(&tokenOpts{
		ConsumerToken: "",
		EmployeeToken: "",
	})

  tt, err := tripletex.New(&tripletex.APIClientOpts{})

  if err != nil {
    return fmt.Errorf("whatever failed: %w", err)
  }

	customersRes, err := c.CustomerSearchWithResponse(context.Background(), &CustomerSearchParams{})
  if err != nil {
    return fmt.Errorf("failed to search for customers: %w", err)
  }

  // Do something with customersRes

  return nil
}
```

## Things to know

- Tripletex's OpenAPI specification is valid, but not error-free.
  - There are duplicate types (eg. `LeaveOfAbsenceType`).
  - No endpoint specifies what the returning content-type is.
  - Emails can returned as empty strings (`""`).
- We convert the original [Tripletex API] from Swagger 2.0 to OpenAPI 3 with
  [Swagger's official tooling](https://converter.swagger.io/api/convert).

[tripletex]: https://tripletex.no
[tripletex api]: https://tripletex.no/v2/swagger.json
[oapi-codegen]: https://github.com/oapi-codegen/oapi-codegen
