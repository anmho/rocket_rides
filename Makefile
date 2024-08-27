
all: build

.PHONY:
build:
	@go build -o ./bin/api ./cmd/api/main.go

.PHONY: run
run: build
	@./bin/api

.PHONY: clean
clean:
	@rm ./bin/*

.PHONY: test
test:
	@go test ./...

.PHONY: check
check:
	@go vet ./...