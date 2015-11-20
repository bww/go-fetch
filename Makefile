
export GOPATH := $(GOPATH):$(PWD)

BIN=bin
PRODUCT=$(BIN)/gofetch

.PHONY: all deps test gofetch install

all: gofetch

deps:

test:

gofetch: deps
	go build -o $(PRODUCT) ./src/gofetch

install: gofetch
	install $(PRODUCT) /usr/local/bin/
