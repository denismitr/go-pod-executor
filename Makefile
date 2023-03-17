.PHONY: deps
deps:
	go mod tidy
	go mod vendor

.PHONY: test
test:
	go test -v ./...