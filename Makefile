src = $(wildcard *.go)

.PHONY: all clean fmt lint megacheck release snapshot test

all: test

clean:
	rm -rf ./dist

coverage:
	find . -iname "*_test.go" -type f | \
		grep -v /vendor/ | \
		sed -E 's|/[^/]+$$||' | \
		xargs -I {} -L 1 go test -coverprofile={}/.coverprofile {}
	gover

fmt:
	find . -iname '*.go' -type f -not -path './vendor/*' -exec gofmt -s -l {} \;
	test -z $(gofmt -s -l $(find . -iname '*.go' -type f | grep -v /vendor/))

cyclo:
	gocyclo -over 19 -avg $$(find . -iname '*.go' -type f | grep -v /vendor/)

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
