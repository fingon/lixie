/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package main

import "strconv"

type FormValued interface {
	FormValue(string) string
}

func intFromForm(r FormValued, key string, value *int) (found bool, err error) {
	raw := r.FormValue(key)
	if raw == "" {
		return
	}
	*value, err = strconv.Atoi(raw)
	return
}

func uint64FromForm(r FormValued, key string, value *uint64) (found bool, err error) {
	raw := r.FormValue(key)
	if raw == "" {
		return
	}
	*value, err = strconv.ParseUint(raw, 10, 64)
	return
}

func boolFromForm(r FormValued, key string, value *bool) {
	raw := r.FormValue(key)
	if raw == "" {
		return
	}
	*value = true
}
