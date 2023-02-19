.PHONY: default clean test race fulltest cover build run install

VERSION=$(shell git describe --always --long --dirty)

default: test

clean:
	rm -rf ./build
	rm -f ./coverage.out
	rm -f ./*.prof

test: clean
	go test -cover ./...
	golangci-lint run

race: clean
	go test -cover -race ./...

fulltest: clean
	go test -cover -race ./...
	golangci-lint run

cover: clean
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

run:
	DEV=1 go run -ldflags="-X main.gitHash=$(VERSION)" ./...

install:
	go install -ldflags="-X main.gitHash=$(VERSION)" ./...

ldupe: install
	DEV=1 fmk --copy d ./testdata/100-dupe/

lsort: install
	rm -rf ./testdata/20/foo
	rm -rf ./testdata/20/bar
	rm -rf ./testdata/20/baz
	DEV=1 fmk --copy s -f "foo,bar,baz" ./testdata/20/

msort: install
	rm -rf ./testdata/20/foo
	rm -rf ./testdata/20/bar
	rm -rf ./testdata/20/baz
	DEV=1 fmk s -f "foo,bar,baz" ./testdata/20/

lss:
	DEV=1 fmk ss ./testdata/sets/

lrename: install
	DEV=1 fmk r ./testdata/20/
