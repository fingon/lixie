/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 * Created:       Sun Jun  2 10:46:21 2024 mstenber
 * Last modified: Sun Jun  2 18:16:15 2024 mstenber
 * Edit time:     6 min
 *
 */

package main

import (
	"encoding/json"
	"io"
	"net/http"
	"runtime/debug"
)

type VersionInfo struct {
	BuildTimestamp string
	BuildInfo      debug.BuildInfo
}

var BuildTimestamp = "not set"

func versionHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bi, _ := debug.ReadBuildInfo()
		w.WriteHeader(http.StatusOK)
		if r.FormValue("simple") != "" {
			_, _ = io.WriteString(w, BuildTimestamp)
			return
		}

		vi := VersionInfo{BuildTimestamp: BuildTimestamp}
		if bi != nil {
			vi.BuildInfo = *bi
		}
		b, err := json.Marshal(vi)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		_, _ = w.Write(b)
	})
}
