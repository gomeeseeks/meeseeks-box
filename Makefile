src = $(wildcard *.go)

.PHONY: all test release clean

all: test

test:
	go test -cover ./...

snapshot: test
	goreleaser --snapshot --rm-dist

release: test
	goreleaser --rm-dist

clean:
	rm -rf ./dist
