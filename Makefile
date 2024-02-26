#
# Author: Markus Stenberg <fingon@iki.fi>
#
# Copyright (c) 2024 Markus Stenberg
#

install-templ:
	go install github.com/a-h/templ/cmd/templ@latest

serve:
	templ generate --watch --proxy="http://localhost:8080" --cmd="go run ."
