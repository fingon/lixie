#
# Author: Markus Stenberg <fingon@iki.fi>
#
# Copyright (c) 2024 Markus Stenberg
#

BINARY=lixie
TEMPLATES = $(wildcard *.templ)
GENERATED = $(patsubst %.templ,%_templ.go,$(TEMPLATES))

build: $(BINARY)

$(BINARY): $(wildcard */*.go) $(wildcard *.go) $(GENERATED)
	go build .
	go test

.PHONY: clean
clean:
	rm -f *_templ.go *_templ.txt $(BINARY)

.PHONY: install-templ
install-templ:
	go install github.com/a-h/templ/cmd/templ@latest

.PHONY: serve
serve:
	templ generate --watch --proxy="http://localhost:8080" --cmd="go run ."


%_templ.go: %.templ
	templ generate -f $<


update-sample:
	wget -r -np -k -l 1 http://localhost:8080/
