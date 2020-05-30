VERSION := $(shell git describe --always --long --dirty)
all: install

fetch:
	go get ./...

install: fetch
	@echo Installing to ${GOPATH}/bin
	go -v install

test: fetch
	@echo Running tests
	go -v test


coverage: fetch
	@echo Running Test with Coverage export
	go test -v -coverprofile=cover.out
	go tool cover -html=cover.out -o cover.html
	#go test -json > report.json
