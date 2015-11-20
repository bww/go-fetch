
export GOPATH := $(GOPATH):$(PWD)

.PHONY: all deps test

all: gofetch

deps:

gofetch: deps
	go build -o bin/gofetch ./src/main
