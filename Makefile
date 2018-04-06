src = $(wildcard *.go)

.PHONY: all clean lint megacheck release snapshot test

all: test

clean:
	rm -rf ./dist

lint:
	golint -set_exit_status $(go list ./...)

megacheck:
	megacheck $(go list ./...)

release: test
	goreleaser --rm-dist

snapshot: test
	goreleaser --snapshot --rm-dist

test: lint megacheck
	go test -cover ./...
