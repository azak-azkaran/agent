VERSION := $(shell git describe --always --long --dirty)
all: install

fetch:
	go mod download

install: fetch
	@echo Installing to ${GOPATH}/bin
	go install -v

run:
	@echo Running agent
	go build -v
	agent

test: fetch
	@echo Running tests
	export RUN_MOCK=true
	go test -v

coverage: fetch
	@echo Running Test with Coverage export
	go test -v -coverprofile=cover.out
	go test -json > report.json
	go tool cover -html=cover.out -o cover.html

coverall: coverage
	@echo Running Test with Coverall
	goveralls -coverprofile=cover.out -service=travis-ci -repotoken=${COVERALLS_TOKEN }
