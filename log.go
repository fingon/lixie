/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/fingon/lixie/data"
)

func logLink(log *data.Log, op string) templ.SafeURL {
	return templ.URL(topLevelLog.Path + fmt.Sprintf("/%d/%s", log.Hash(), op))
}

func toJson(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

func logClassifyHandler(db *data.Database, ham bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hash_string := r.PathValue("hash")
		hash, err := strconv.ParseUint(hash_string, 10, 64)
		if err != nil {
			// TODO handle error
			return
		}
		if db.ClassifyHash(hash, ham) {
			http.Redirect(w, r, topLevelLog.Path, http.StatusSeeOther)
			return
		}
		http.NotFound(w, r)
	})
}
