/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 * Created:       Fri Apr 26 10:44:18 2024 mstenber
 * Last modified: Thu Oct 31 08:02:09 2024 mstenber
 * Edit time:     42 min
 *
 */

package cm

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"gotest.tools/v3/assert"
)

type empty struct{}

type tt struct {
	B  bool   `json:"b" cm:"bf"`
	I  int64  `json:"i" cm:"if"`
	II int    `json:"ii" cm:"iif"`
	S  string `json:"s" cm:"sf"`
	U  uint64 `json:"u" cm:"uf"`
}

type staticCookie struct {
	// FormValued interface provider
	URLWrapper

	// Data for rest of CookieSource
	name   string
	cookie *http.Cookie
	err    error
}

func (self *staticCookie) Cookie(name string) (*http.Cookie, error) {
	if name != self.name {
		return nil, fmt.Errorf("Invalid name: %s <> %s", name, self.name)
	}
	if self.cookie == nil {
		return nil, http.ErrNoCookie
	}
	return self.cookie, self.err
}

func TestParse(t *testing.T) {
	// first off, ensure the type checking is correct. anything
	// else than pointer (to a struct) shouldn't work
	_, err := Parse(nil, nil, nil)
	assert.Assert(t, err, nil)

	es := empty{}
	_, err = Parse(nil, nil, es)
	assert.Assert(t, err, nil)

	sc0 := staticCookie{name: "cm-cm.empty"}

	i := 42
	_, err = Parse(&sc0, &sc0.URLWrapper, &i)
	assert.Assert(t, err != nil)

	_, err = Parse(&sc0, &sc0.URLWrapper, &es)
	assert.Equal(t, err, nil)

	// empty is fine but doesn't do anything yet
	sc1 := staticCookie{name: "cm-cm.empty"}
	changed, err := Parse(&sc1, &sc1.URLWrapper, &es)
	assert.Equal(t, err, nil)
	assert.Equal(t, changed, false)

	q1, err := ToQueryString("x", &sc1)
	assert.Equal(t, err, nil)
	assert.Equal(t, q1, "x")

	ts := tt{}
	sc2 := staticCookie{name: "cm-cm.tt"}
	changed, err = Parse(&sc2, &sc2.URLWrapper, &ts)
	assert.Equal(t, err, nil)
	assert.Equal(t, changed, false)

	sc3 := staticCookie{name: "cm-cm.tt", URLWrapper: URLWrapper(url.Values{
		"bf":  []string{"true"},
		"if":  []string{"-1"},
		"sf":  []string{"str"},
		"iif": []string{"7"},
		"uf":  []string{"42"},
	})}
	assert.Equal(t, sc3.FormValue("bf"), "true")
	changed, err = Parse(&sc3, &sc3.URLWrapper, &ts)
	assert.Equal(t, err, nil)
	assert.Equal(t, changed, true)
	assert.Equal(t, ts.B, true)
	assert.Equal(t, ts.I, int64(-1))
	assert.Equal(t, ts.U, uint64(42))

	q3, err := ToQueryString("x", &ts)
	assert.Equal(t, err, nil)
	assert.Equal(t, q3, "x?bf=true&if=-1&iif=7&sf=str&uf=42")

	// Export the cookie
	cookie, err := ToCookie(&ts)
	assert.Equal(t, err, nil)
	assert.Assert(t, cookie != nil)

	// Pretend we're new request: take in the cookie
	sc4 := staticCookie{name: "cm-cm.tt", cookie: cookie, err: nil}
	ts4 := tt{}

	changed, err = Parse(&sc4, &sc4.URLWrapper, &ts4)
	assert.Equal(t, changed, false)
	assert.Equal(t, err, nil)

	// The resulting query string (=~ content) should be same
	q4, err := ToQueryString("x", &ts4)
	assert.Equal(t, err, nil)
	assert.Equal(t, q3, q4)
}
