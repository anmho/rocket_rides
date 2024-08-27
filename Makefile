
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

.PHONY: lines
lines:
	@git log --author="Andrew Ho" --pretty=tformat: --numstat \
     | gawk '{ add += $$1; subs += $$2; loc += $$1 - $$2 } END { printf "added lines: %s removed lines: %s total lines: %s\n", add, subs, loc }' -



