export CGO_ENABLED ?= 0

VERSION ?= $(shell git describe --tags 2>/dev/null | sed -e 's/^v//g' || echo "dev")
REVISION ?= $(shell git rev-parse --short HEAD || echo "unknown")
DATE ?= $(shell date +%Y-%m-%dT%H:%M:%S%:z)
LDFLAGS = -s -w \
          -X github.com/gomeeseeks/meeseeks-box/version.Version=$(VERSION) \
          -X github.com/gomeeseeks/meeseeks-box/version.Commit=$(REVISION) \
          -X github.com/gomeeseeks/meeseeks-box/version.Date=$(DATE)

src = $(wildcard *.go)

.PHONY: all test compile release clean build_docker

all: test

test:
	go test -cover ./...

compile:
	go build -o meeseeks-box -ldflags "$(LDFLAGS)" main.go

snapshot: test
	goreleaser --snapshot --rm-dist

release: test
	goreleaser --rm-dist

clean:
	rm -rf ./dist

CI_REGISTRY_IMAGE ?= gomeeseeks/meeseeks-box

build_docker:
	docker build -t $(CI_REGISTRY_IMAGE):$(VERSION) -f Dockerfile .
	if [ -n "$$CI_REGISTRY" ]; then \
	  docker login --username $$CI_REGISTRY_USER --password $$CI_REGISTRY_PASSWORD $$CI_REGISTRY && \
	  docker push $(CI_REGISTRY_IMAGE):$(VERSION) && \
	  docker logout $$CI_REGISTRY; \
	fi
