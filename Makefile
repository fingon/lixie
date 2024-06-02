#
# Author: Markus Stenberg <fingon@iki.fi>
#
# Copyright (c) 2024 Markus Stenberg
#

BINARY=lixie
TEMPLATES = $(wildcard *.templ)
GENERATED = $(patsubst %.templ,%_templ.go,$(TEMPLATES))
TEMPL_VERSION = $(shell grep a-h/templ go.mod | sed 's/^.* v/v/')
BUILD_TIMESTAMP=$(shell date "+%Y-%m-%dT%H:%M:%S")

all: build lint

build: $(BINARY)

# See https://golangci-lint.run/usage/linters/
lint:
	golangci-lint run --fix  # Externally installed, e.g. brew

fmt:
	go fmt ./...
	templ fmt .

$(BINARY): $(wildcard */*.go) $(wildcard *.go) $(GENERATED) Makefile
	go test ./... -race -covermode=atomic -coverprofile=coverage.out
	go build -ldflags="-X main.BuildTimestamp=$(BUILD_TIMESTAMP)" .

.PHONY: clean
clean:
	rm -f *_templ.go *_templ.txt $(BINARY)

.PHONY:
dep:
	go install github.com/a-h/templ/cmd/templ@$(TEMPL_VERSION)

.PHONY: serve
serve: run
	watchman-make -p '**/*.go' '**/*.templ' -t run
# templ generate --watch --proxy="http://localhost:8080" --cmd="go run ."
# ^ sometimes bugs, which is unfortunate

.PHONY: run
run: $(BINARY)
	killall -9 -q lixie && sleep 1 || true
	./lixie &

%_templ.go: %.templ
	templ generate -f $<

upgrade:
	go get -u ./...
	go mod tidy

# This is unlikely to work for anyone else than me, but..
update-sample:
	rm -rf ./localhost:8080
	wget -r -np -k -l 1 http://localhost:8080/
	rsync -a --delete \
		./localhost:8080/ ~/sites/fingon.kapsi.fi/www/lixie/
	cd ~/sites && ./update.sh
