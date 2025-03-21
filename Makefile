openapi_path = "./api/openapi.json"
tmp_path = "./tmp.json"

.PHONY: tidy
tidy:
	go mod tidy
	go fmt ./...

.PHONY: generate
generate:
	go generate ./...
	go fmt ./...
	
.PHONY: test
test:
	op run --no-masking --env-file='./.env' -- go test -short ./...

.PHONY: test-watch
test-watch:
	op run --no-masking --env-file='./.env' --  watch -n 5 go test -short ./...

.PHONY: test-full
test-full:
	op run --no-masking --env-file='./.env' -- go test -v ./...

.PHONY: api-openapi-cleanup
api-openapi-cleanup:
	@echo "openapi: overriding bad type name"
	@sed -i '' 's/"LeaveOfAbsenceType":{"required":\["leaveOfAbsenceType"\],"type":"object"/"LeaveOfAbsenceType":{"required":\["leaveOfAbsenceType"\],"type":"object","x-go-name":"LeaveOfAbsenceTypeType"/g' $(tmp_path)
	@sed -i '' 's|\*/\*|application/json|g' $(tmp_path)
	@sed -i '' 's/,"format":"email"//g' $(tmp_path)
	@sed -i '' 's/TimesheetEntry_search/TimesheetEntry_search_search/g' $(tmp_path)
	@cat $(tmp_path) | jq "." > $(openapi_path)
	
.PHONY: api-openapi
api-openapi:
	@echo "openapi: getting latest Swagger from Tripletex and converting to OpenAPI 3"
	@curl https://converter.swagger.io/api/convert?url=https://tripletex.no/v2/swagger.json > $(tmp_path)
	api-openapi-cleanup

.PHONY: check
check:
	go tool github.com/golangci/golangci-lint/cmd/golangci-lint run ./...
	go tool honnef.co/go/tools/cmd/staticcheck ./...
