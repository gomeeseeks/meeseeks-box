src = $(wildcard *.go)
executable        = meeseeks-box
namespace         = pcarranza
version           = ${CI_COMMIT_TAG}

.PHONY: all version test package build build-linux build-darwin build-arm package package-linux packager-arm release release-linux release-arm clean

all: test version build package release clean

version:
	if [ ! -z "$(version)" ]; then \
		sed -i "" -e 's/"unset"/"$(version)"/g' version/version.go ; \
	fi

test:
	go test -cover ./...

build-linux: test version
	mkdir -p build/linux
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-w' -o build/linux/$(executable)

build-arm: test version
	mkdir -p build/arm
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -a -ldflags '-w' -o build/arm/$(executable)

build-darwin: test version
	mkdir -p build/darwin
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -ldflags '-w' -o build/darwin/$(executable)

build: test build-linux build-arm build-darwin

package-linux: build-linx
	cp Dockerfile build/linux && \
		cd build/linux && \
		docker build . -t $(namespace)/$(executable):latest

package-arm: build-arm
	cp Dockerfile build/arm && \
		cd build/arm && \
		sed -i '' -e 's/alpine/arm32v6\/alpine/g' Dockerfile && \
		docker build . -t $(namespace)/$(executable)-armv6:latest

package: build package-linux package-arm

release-linux: package-linux
	docker push $(namespace)/$(executable):latest
	if [ ! -z "$(version)" ]; then \
		docker tag $(namespace)/$(executable):latest $(namespace)/$(executable):$(version) ; \
		docker push $(namespace)/$(executable):$(version) ; \
	fi

release-arm: package-arm
	docker push $(namespace)/$(executable)-armv6:latest
	if [ ! -z "$(version)" ]; then \
		echo "Has version on release arm $(version)" ; \
		docker tag $(namespace)/$(executable)-armv6:latest $(namespace)/$(executable)-armv6:$(version) ; \
		docker push $(namespace)/$(executable)-armv6:$(version) ; \
	fi

release: release-linux

clean:
	rm -rf ./build
	git reset --hard
