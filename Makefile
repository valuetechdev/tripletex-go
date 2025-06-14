openapi_path = "./api/openapi.json"

.PHONY: generate
generate:
	go generate ./...
	go fmt ./...
	
.PHONY: api-openapi
api-openapi:
	@echo "openapi: getting latest Swagger from Tripletex and converting to OpenAPI 3"
	@curl https://converter.swagger.io/api/convert?url=https://tripletex.no/v2/swagger.json > $(openapi_path)
	@echo "openapi: fixing content types"
	@sed -i '' 's|\*/\*|application/json|g' $(openapi_path)

.PHONY: check
check:
	go tool github.com/golangci/golangci-lint/cmd/golangci-lint run ./...
	go tool honnef.co/go/tools/cmd/staticcheck ./...
