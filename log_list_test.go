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

	"github.com/fingon/lixie/cm"
	"gotest.tools/v3/assert"
)

func TestLogList(t *testing.T) {
	conf := LogListConfig{Expand: uint64(42), BeforeHash: 13}
	s := conf.Query().ToLinkString()
	u, err := url.Parse(s)
	assert.Equal(t, err, nil)

	conf1 := LogListConfig{}
	wr := cm.URLWrapper(u.Query())
	err = conf1.Init(nil, &wr, nil)
	assert.Equal(t, err, nil)
	assert.Equal(t, conf, conf1)
}
