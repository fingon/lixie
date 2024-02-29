/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package main

import (
	"net/url"
	"strconv"

	"github.com/a-h/templ"
)

type LogListConfig struct {
	AutoRefresh bool
	Expand      uint64
}

type FormValued interface {
	FormValue(string) string
}

const logListBase = "/log/"

const autoRefreshKey = "ar"
const expandKey = "exp"

func NewLogListConfig(r FormValued) LogListConfig {
	autorefresh := r.FormValue(autoRefreshKey) != ""
	expand := uint64(0)
	sv := r.FormValue(expandKey)
	if sv != "" {
		v, err := strconv.ParseUint(sv, 10, 64)
		if err == nil {
			expand = v
		}
	}
	return LogListConfig{AutoRefresh: autorefresh, Expand: expand}
}

func (self LogListConfig) WithAutoRefresh(v bool) LogListConfig {
	self.AutoRefresh = v
	return self
}

func (self LogListConfig) WithExpand(v uint64) LogListConfig {
	self.Expand = v
	return self
}

func (self LogListConfig) ToLinkString() string {
	// TODO should this be constant or parameter?
	base := logListBase
	v := url.Values{}
	if self.AutoRefresh {
		v.Set(autoRefreshKey, "1")
	}
	if self.Expand != 0 {
		v.Set(expandKey, strconv.FormatUint(self.Expand, 10))
	}
	if len(v) == 0 {
		return base
	}
	return base + "?" + v.Encode()
}

func (self LogListConfig) ToLink() templ.SafeURL {
	return templ.URL(self.ToLinkString())
}
