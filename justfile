default: check test

check:
    gofumpt -l . | tee /dev/stderr | (! grep .)
    go vet ./...
    golangci-lint run

fmt:
    gofumpt -l -w .

test:
    go test -race ./...

tidy:
    go mod tidy
