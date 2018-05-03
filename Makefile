GOPATH=$(CURDIR)

IS_GCCGO_INSTALLED=$(gccgo --version 2> /dev/null)

# build version
VERSION=`git describe --tags`
BUILD_NUMBER=`git rev-parse HEAD`
BUILD_DATE=`date +%Y-%m-%d-%H:%M`

# go compiler flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildNumber=$(BUILD_NUMBER) -X main.BuildDate=$(BUILD_DATE)"
LDFLAGS_COMPRESSED=-ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildNumber=$(BUILD_NUMBER) -X main.BuildDate=$(BUILD_DATE)"

#gccgo compiler flags
GCCGOFLAGS=-gccgoflags "-march=native -O3"
GCCGOFLAGS_GOLD=-gccgoflags "-march=native -O3 -fuse-ld=gold"

# default task
default: build

# install dependencies
deps:
	GOPATH=$(GOPATH) go get -d ./...

deps-gccgo:
ifndef IS_GCCGO_INSTALLED
        $(error "gccgo not installed")
endif

# build with go compiler
build: deps
	GOPATH=$(GOPATH) CGO_ENABLED=0 go build -a $(LDFLAGS) -o $(CURDIR)/bin/statsd-http-proxy


# build with go compiler and link optiomizations
build-shrink: deps
	GOPATH=$(GOPATH) CGO_ENABLED=0 go build -a $(LDFLAGS_COMPRESSED) -o $(CURDIR)/bin/statsd-http-proxy-shrink

# build with gccgo compiler
# Require to install gccgo
build-gccgo: deps deps-gccgo
	GOPATH=$(GOPATH) CGO_ENABLED=0 go build -a -compiler gccgo $(GCCGOFLAGS) -o $(CURDIR)/bin/statsd-http-proxy-gccgo

# build with gccgo compiler and gold linker
# Require to install gccgo
build-gccgo-gold: deps deps-gccgo
	GOPATH=$(GOPATH) CGO_ENABLED=0 go build -a -compiler gccgo $(GCCGOFLAGS_GOLD) -o $(CURDIR)/bin/statsd-http-proxy-gccgo-gold

# build all
build-all: build build-shrink build-gccgo build-gccgo-gold

# clean build
clean:
	rm -rf ./bin
	go clean
	
# publish docker to hub
publish:
	docker build --tag sokil/statsd-http-proxy:latest -f ./Dockerfile.alpine .
	docker push sokil/statsd-http-proxy:latest
	docker build --tag sokil/statsd-http-proxy:$(VERSION) -f ./Dockerfile.alpine .
	docker push sokil/statsd-http-proxy:$(VERSION)
