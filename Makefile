.PHONY: run-tests
run-tests:
	go test ./...

.PHONY: coverage
coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html

.PHONY: lint
lint:
	go fmt ./...
	go mod tidy
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run --timeout 5m

.PHONY: scan-secrets
scan-secrets:
	detect-secrets scan \
		--exclude-files '^tests/.*' \
		> .secrets.baseline
	detect-secrets audit .secrets.baseline
