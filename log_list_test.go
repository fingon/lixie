/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package main

import (
	"net/url"
	"testing"

	"github.com/magiconair/properties/assert"
)

type URLWrapper url.Values

func (self URLWrapper) FormValue(k string) string {
	l, ok := self[k]
	if !ok {
		return ""
	}
	for _, v := range l {
		return v
	}
	return ""
}

func TestLogList(t *testing.T) {
	conf := LogListConfig{Expand: uint64(42), AutoRefresh: true}
	s := conf.ToLinkString()
	u, err := url.Parse(s)
	assert.Equal(t, err, nil)

	conf2 := NewLogListConfig(URLWrapper(u.Query()))
	assert.Equal(t, conf, conf2)
}
