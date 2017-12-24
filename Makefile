src = $(wildcard *.go)
executable        = meeseeks-box
namespace         = pcarranza
version           = 0.0.1

.PHONY: all package build package release

all: test build package release deploy

test:
	go vet -v
	go test -v ./...

build-linux: test
	mkdir -p build/linux
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-w' -o build/linux/$(executable)

build-arm: test
	mkdir -p build/arm
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -a -ldflags '-w' -o build/arm/$(executable)

build-darwin: test
	mkdir -p build/darwin
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -ldflags '-w' -o build/darwin/$(executable)

build: test build-linux build-arm build-darwin

package-linux: build-linux
	cp Dockerfile build/linux
	cd build/linux && \
		docker build . -t $(namespace)/$(executable):latest && \
		docker tag $(namespace)/$(executable):latest $(namespace)/$(executable):$(version)

package-arm: build-arm
	cp Dockerfile build/arm && \
		cd build/arm && \
		sed -i '' -e 's/alpine/arm32v6\/alpine/g' Dockerfile && \
		docker build . -t $(namespace)/$(executable)-armv6:latest && \
		docker tag $(namespace)/$(executable)-armv6:latest $(namespace)/$(executable)-armv6:$(version)

package-darwin: build-darwin
	cp Dockerfile build/darwin && \
		cd build/darwin	&& \
		docker build . -t $(namespace)/$(executable)-darwin:latest && \
		docker tag $(namespace)/$(executable)-darwin:latest $(namespace)/$(executable)-darwin:$(version)

package: build package-linux

release-linux: package-linux
	docker push $(namespace)/$(executable):latest
	docker push $(namespace)/$(executable):$(version)

release-arm: package-arm
	docker push $(namespace)/$(executable)-armv6:latest
	docker push $(namespace)/$(executable)-armv6:$(version)

release: package release-linux release-arm
