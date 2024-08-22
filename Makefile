
all: build

.PHONY:
build:
	@go build -o ./bin/server main.go

.PHONY: run
run:
	@./bin/server