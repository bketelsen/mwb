.PHONY: build install clean test fmt lint check

build:
	go build -o mwb ./cmd/mwb

install:
	go install ./cmd/mwb

clean:
	rm -f mwb

test:
	go test ./...

fmt:
	gofmt -w .

lint:
	golangci-lint run

check: fmt lint test
