/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package cm

import "strconv"

type FormValued interface {
	FormValue(string) string
}

func IntFromForm(r FormValued, key string, value *int) (found bool, err error) {
	raw := r.FormValue(key)
	if raw == "" {
		return
	}
	found = true
	*value, err = strconv.Atoi(raw)
	return
}

func Int64FromForm(r FormValued, key string, value *int64) (found bool, err error) {
	raw := r.FormValue(key)
	if raw == "" {
		return
	}
	found = true
	*value, err = strconv.ParseInt(raw, 10, 64)
	return
}

func Uint64FromForm(r FormValued, key string, value *uint64) (found bool, err error) {
	raw := r.FormValue(key)
	if raw == "" {
		return
	}
	found = true
	*value, err = strconv.ParseUint(raw, 10, 64)
	return
}

func BoolFromForm(r FormValued, key string, value *bool) {
	raw := r.FormValue(key)
	if raw == "" || raw == "false" {
		return
	}
	*value = true
}
