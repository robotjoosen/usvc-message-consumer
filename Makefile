.PHONY: lint
lint:
	go mod tidy
	gci write --skip-generated . && gofumpt -l -w .
	golangci-lint run --fix -c .golangci.yml ./... -v
