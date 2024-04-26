/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 * Created:       Fri Apr 26 11:32:12 2024 mstenber
 * Last modified: Fri Apr 26 14:33:25 2024 mstenber
 * Edit time:     0 min
 *
 */

package cm

import "net/url"

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

func (self URLWrapper) HasValue(k string) bool {
	_, ok := self[k]
	return ok
}
