all: test

.PHONY: test
test:
	go test -race -coverprofile=coverage.txt -covermode=atomic

.PHONY: cover
cover: test
	go tool cover -html=coverage.txt
