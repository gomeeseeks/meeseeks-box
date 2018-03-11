src = $(wildcard *.go)

.PHONY: all clean lint release snapshot test

all: test

clean:
	rm -rf ./dist

lint:
	golint -set_exit_status ./...

release: test
	goreleaser --rm-dist

snapshot: test
	goreleaser --snapshot --rm-dist

test: lint
	go test -cover ./...
