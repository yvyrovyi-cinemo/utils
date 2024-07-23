
.PHONY: deps
deps:
	go mod tidy
	go mod download

.PHONY: lilnt
lint:
	golangci-lint run -v -c ./.golangci.yml
