VERSION := $(shell git describe --always --long --dirty)
all: install

fetch:
	go get ./...

install: fetch
	@echo Installing to ${GOPATH}/bin
	go install -v

test: fetch
	@echo Running tests
	export RUN_MOCK=true
	go test -v


coverage: fetch
	@echo Running Test with Coverage export
	go test -v -coverprofile=cover.out
	go tool cover -html=cover.out -o cover.html
