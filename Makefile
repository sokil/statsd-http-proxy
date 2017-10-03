GOPATH=$(CURDIR)

default: build

# install dependencies
deps:
	GOPATH=$(GOPATH) go get -d ./...

# build with go compiler
build: deps
	GOPATH=$(GOPATH) go build -o $(CURDIR)/bin/statsd -o $(CURDIR)/bin/statsd


# build with go compiler with links optiomizations
build-shrink: deps
	GOPATH=$(GOPATH) go build -ldflags="-s -w" -o $(CURDIR)/bin/statsd-shrink

# build with gccgo compiler
# Require to install gccgo
build-gccgo: deps
	GOPATH=$(GOPATH) go build -compiler gccgo -gccgoflags "-march=native -O3" -o $(CURDIR)/bin/statsd-gccgo

# build with gccgo compiler with gold linker
# Require to install gccgo
build-gccgo-gold: deps
	GOPATH=$(GOPATH) go build -compiler gccgo -gccgoflags "-march=native -O3 -fuse-ld=gold" -o $(CURDIR)/bin/statsd-gccgo-gold

# clean build
clean:
	rm -rf ./bin
	rm statsd-rest-server
	go clean
