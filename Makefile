.PHONY: build test vet coverage

build:
	go build ./...

test:
	go test ./...

vet:
	go vet ./...

coverage:
	go test -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report written to coverage.html"
