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

all: fmt build lint

build: $(BINARY)

# See https://golangci-lint.run/usage/linters/
lint:
	golangci-lint run --fix  # Externally installed, e.g. brew

fmt:
	go fmt ./...
	templ fmt .

$(BINARY): $(wildcard */*.go) $(wildcard *.go) $(GENERATED) Makefile
	go test ./... \
		-race \
		-coverpkg=./... -covermode=atomic -coverprofile=coverage.out
	go build -ldflags="-X main.ldBuildTimestamp=$(BUILD_TIMESTAMP)" .

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
	killall -9 -q lixie || true
	./lixie -dev &

%_templ.go: %.templ
	templ generate -f $<

upgrade:
	go get -u ./...
	go mod tidy

# This is unlikely to work for anyone else than me, but..
update-sample:
	killall -9 -q lixie || true
	./lixie &
	sleep 3
	rm -rf ./localhost:8080
	wget -r -np -k -l 1 http://localhost:8080/
	rsync -a --delete \
		./localhost:8080/ ~/sites/fingon.kapsi.fi/www/lixie/
	cd ~/sites && ./update.sh

validate-codecov:
	curl --data-binary @codecov.yml https://codecov.io/validate

.venv: Makefile .venv/bin/activate

.venv/bin/activate: $(wildcard requirements*.txt)
	rm -rf .venv
	uv venv
	uv pip install -r requirements.txt
