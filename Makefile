.PHONY: build install clean test

build:
	go build -o mwb ./cmd/mwb

install:
	go install ./cmd/mwb

clean:
	rm -f mwb

test:
	go test ./...
