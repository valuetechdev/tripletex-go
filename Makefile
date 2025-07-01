openapi_path = "./api/openapi.json"

.PHONY: generate
generate:
	go generate ./...
	go fmt ./...
	
.PHONY: api-openapi
api-openapi:
	@curl https://tripletex.no/v2/openapi.json | jq -c > $(openapi_path)
	@echo "openapi: fixing content types"
	@sed -i '' 's|\*/\*|application/json|g' $(openapi_path)

.PHONY: check
check:
	go tool github.com/golangci/golangci-lint/cmd/golangci-lint run ./...
	go tool honnef.co/go/tools/cmd/staticcheck ./...
