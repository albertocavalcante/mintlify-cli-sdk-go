default: check

build:
    go build ./...

test:
    go test ./...

test-v:
    go test -v ./...

lint:
    go vet ./...

check: build lint test

fmt:
    gofmt -w .
