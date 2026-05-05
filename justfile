default: check test

check: modernize
    gofumpt -l . | tee /dev/stderr | (! grep .)
    go vet ./...
    golangci-lint run

# Apply Go modernizations (min/max, range over int, new(value), etc.)
# and fail if the working tree is no longer clean against HEAD for *.go.
# This makes modernize signal mandatory rather than advisory.
modernize:
    go fix ./...
    @if ! git diff --quiet -- '*.go'; then \
        echo ""; \
        echo "ERROR: go fix ./... produced changes. Review with 'git diff' and commit before continuing."; \
        echo "(If these changes are unrelated work-in-progress, stage them and re-run.)"; \
        git diff --stat -- '*.go'; \
        exit 1; \
    fi

fmt:
    gofumpt -l -w -extra .

test:
    go test -race ./...

tidy:
    go mod tidy
