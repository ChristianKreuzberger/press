.PHONY: build test vet coverage

build:
	go build ./...

test:
	go test ./...

vet:
	go vet ./...

# Unit coverage plus coverage of the press binary exercised by the e2e tests
# (see TestMain), merged into one profile.
coverage:
	rm -rf covdata && mkdir covdata
	PRESS_E2E_COVERDIR=$(CURDIR)/covdata go test -coverprofile=coverage-unit.out -covermode=atomic ./...
	go tool covdata textfmt -i=covdata -o=coverage-e2e.out
	{ cat coverage-unit.out; grep -v '^mode:' coverage-e2e.out; } > coverage.out
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report written to coverage.html"
