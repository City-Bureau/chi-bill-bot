.PHONY: install test format lint build deploy clean

functions := $(shell find functions -name \*main.go | awk -F'/' '{print $$2}')

install:
	npm install

test:
	go test ./...

format:
	test -z $$(gofmt -l .)

lint:
	golangci-lint run

build:
	@for function in $(functions) ; do \
		env GOOS=linux go build -ldflags="-s -w" -o bin/$$function functions/$$function/main.go ; \
	done

deploy:
	npx serverless deploy --stage prod

clean:
	rm -rf bin
