/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 * Created:       Fri Apr 26 14:09:03 2024 mstenber
 * Last modified: Fri Apr 26 14:40:37 2024 mstenber
 * Edit time:     5 min
 *
 */

package cm

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
)

func Parse(r CookieSource, u URLWrapper, state any) (changed bool, err error) {
	if r == nil {
		err = ErrNoSource
		return
	}
	name, err := cookieName(state)
	if err != nil {
		return
	}

	// Handle the cookie, if any
	cookie, err := r.Cookie(name)
	switch {
	case err == http.ErrNoCookie:
		// fmt.Printf("Cookie not found\n")
		err = nil
	case err != nil:
		// fmt.Printf("Failed to get cookie: %s", err)
		return
	case cookie != nil:
		err = cookie.Valid()
		if err != nil {
			return
		}
		data, err := base64.StdEncoding.DecodeString(cookie.Value)
		if err != nil {
			return false, err
		}
		err = json.Unmarshal(data, state)
		if err != nil {
			return false, err
		}
	}

	// Struct we're filling from the form
	s := reflect.ValueOf(state).Elem()

	// Go through the fields and fill in the values IF they differ from what is in the struct
	for _, field := range reflect.VisibleFields(s.Type()) {
		formKey := field.Tag.Get("cm")
		if formKey == "" {
			continue
		}
		if !u.HasValue(formKey) {
			continue
		}

		value := u.FormValue(formKey)

		f := s.FieldByName(field.Name)
		if !(f.IsValid() && f.CanSet()) {
			err = fmt.Errorf("invalid field %s", field.Name)
			return
		}

		kind := field.Type.Kind()
		switch {
		case kind == reflect.String:
			if f.String() != value {
				f.SetString(value)
				changed = true
			}
		case kind == reflect.Bool:
			var b bool
			BoolFromForm(u, formKey, &b)
			if f.Bool() != b {
				f.SetBool(b)
				changed = true
			}

		case f.CanInt():
			var i int64
			_, err = Int64FromForm(u, formKey, &i)
			if err != nil {
				return
			}
			if f.Int() != i {
				f.SetInt(i)
				changed = true
			}

		case f.CanUint():
			var uv uint64
			_, err = Uint64FromForm(u, formKey, &uv)
			if err != nil {
				return
			}
			if f.Uint() != uv {
				f.SetUint(uv)
				changed = true
			}
		default:
			err = fmt.Errorf("insupported field %s kind: %s", field.Name, kind)
		}
		if err != nil {
			return
		}
	}

	return
}
