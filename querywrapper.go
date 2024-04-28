/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 * Created:       Sun Apr 28 10:03:22 2024 mstenber
 * Last modified: Sun Apr 28 10:25:19 2024 mstenber
 * Edit time:     9 min
 *
 */

package main

import (
	"net/url"

	"github.com/a-h/templ"
)

type QueryWrapper struct {
	Base   string
	Values url.Values
}

func (self *QueryWrapper) Add(key, value string) *QueryWrapper {
	if self.Values == nil {
		self.Values = url.Values{}
	}
	self.Values.Add(key, value)
	return self
}

func (self QueryWrapper) ToLink() templ.SafeURL {
	return templ.URL(self.ToLinkString())
}

func (self QueryWrapper) ToLinkString() string {
	if len(self.Values) > 0 {
		return self.Base + "?" + self.Values.Encode()
	}
	return self.Base
}
