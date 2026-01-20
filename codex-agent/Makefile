.PHONY: build-go build-ts build clean

build: build-ts build-go

build-ts:
	npm run build

build-go:
	cd cmd/implement-ticket && go build -o ../../bin/implement-ticket

clean:
	rm -rf dist/ bin/
	cd cmd/implement-ticket && go clean

install-go: build-go
	cp bin/implement-ticket /usr/local/bin/implement-ticket

.DEFAULT_GOAL := build
