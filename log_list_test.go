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

	"gotest.tools/v3/assert"
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
	conf := LogListConfig{Expand: uint64(42), AutoRefresh: true, Filter: 7}
	s := conf.ToLinkString()
	u, err := url.Parse(s)
	assert.Equal(t, err, nil)

	conf1 := LogListConfig{}
	err = conf1.Init(URLWrapper(u.Query()))
	assert.Equal(t, err, nil)
	assert.Equal(t, conf, conf1)

	// Ensure that sticking in Before means it actually won't be
	// the same after parsing (it shouldn't be parsed here)
	conf.Before = 13
	s2 := conf.ToLinkString()

	assert.Assert(t, s != s2)

	u, err = url.Parse(s2)
	assert.Equal(t, err, nil)

	conf2 := LogListConfig{}
	err = conf2.Init(URLWrapper(u.Query()))
	assert.Equal(t, err, nil)
	assert.Equal(t, conf1, conf2)
}
