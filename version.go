/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 * Created:       Sun Jun  2 10:46:21 2024 mstenber
 * Last modified: Sun Jun  2 10:47:10 2024 mstenber
 * Edit time:     1 min
 *
 */

package main

import (
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
)

var BuildTimestamp = "not set"

func versionHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bi, ok := debug.ReadBuildInfo()
		w.WriteHeader(http.StatusOK)
		if r.FormValue("simple") != "" {
			_, _ = io.WriteString(w, BuildTimestamp)
			return
		}
		_, _ = io.WriteString(w, fmt.Sprintf("Build timestamp: %s\n", BuildTimestamp))
		if !ok {
			_, _ = io.WriteString(w, "No build info available\n")

			return
		}
		for _, setting := range bi.Settings {
			_, _ = io.WriteString(w, fmt.Sprintf("%s=%s\n", setting.Key, setting.Value))
		}
	})
}
