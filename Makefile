install:
	@go get -v golang.org/x/lint/golint
.PHONY: install

test:
	@golint -set_exit_status ./...
	@go test -v -timeout 30s -race -coverprofile=coverage.txt -covermode=atomic ./...
.PHONY: test
