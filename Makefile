GOPATH=$(CURDIR)

# cross-compile rules: https://golang.org/doc/install/source

default: build

# install dependencies
deps:
	GOPATH=$(GOPATH) go get -d ./...

# build with go compiler
build: deps
	GOPATH=$(GOPATH) go build -a -o $(CURDIR)/bin/statsd -o $(CURDIR)/bin/statsd-rest-server


# build with go compiler and link optiomizations
build-shrink: deps
	GOPATH=$(GOPATH) go build -a -ldflags="-s -w" -o $(CURDIR)/bin/statsd-rest-server-shrink

# build with gccgo compiler
# Require to install gccgo
build-gccgo: deps
	GOPATH=$(GOPATH) go build -a -compiler gccgo -gccgoflags "-march=native -O3" -o $(CURDIR)/bin/statsd-rest-server-gccgo

# build with gccgo compiler and gold linker
# Require to install gccgo
build-gccgo-gold: deps
	GOPATH=$(GOPATH) go build -a -compiler gccgo -gccgoflags "-march=native -O3 -fuse-ld=gold" -o $(CURDIR)/bin/statsd-rest-server-gccgo-gold

# clean build
clean:
	rm -rf ./bin
	rm statsd-rest-server
	go clean
