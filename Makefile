functions := $(shell find functions -name \*main.go | awk -F'/' '{print $$2}')

test:
	go test ./...

build:
	@for function in $(functions) ; do \
		env GOOS=linux go build -ldflags="-s -w" -o bin/$$function functions/$$function/main.go ; \
	done

clean:
	rm -rf bin
