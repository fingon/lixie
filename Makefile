#
# Author: Markus Stenberg <fingon@iki.fi>
#
# Copyright (c) 2024 Markus Stenberg
#

BINARY=lixie

build: $(BINARY)

# Note: This is somewhat lazy - we don't try to be smart about when to
# generate something
$(BINARY): $(wildcard *.go) $(wildcard *.templ)
	templ generate .
	go build .

clean:
	rm -f *_templ.go *_templ.txt $(BINARY)

install-templ:
	go install github.com/a-h/templ/cmd/templ@latest

serve:
	templ generate --watch --proxy="http://localhost:8080" --cmd="go run ."
