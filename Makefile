#
# Author: Markus Stenberg <fingon@iki.fi>
#
# Copyright (c) 2024 Markus Stenberg
#

BINARY=lixie
TEMPLATES = $(wildcard *.templ)
GENERATED = $(patsubst %.templ,%_templ.go,$(TEMPLATES))

all: build lint

build: $(BINARY)

# See https://golangci-lint.run/usage/linters/
lint:
	golangci-lint run --fix  # Externally installed, e.g. brew

fmt:
	templ fmt .

$(BINARY): $(wildcard */*.go) $(wildcard *.go) $(GENERATED) Makefile
	go test ./...
	go build .

.PHONY: clean
clean:
	rm -f *_templ.go *_templ.txt $(BINARY)

.PHONY:
dep:
	go install github.com/a-h/templ/cmd/templ@latest

.PHONY: serve
serve:
	templ generate --watch --proxy="http://localhost:8080" --cmd="go run ."


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
