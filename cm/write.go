/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 * Created:       Fri Apr 26 14:08:31 2024 mstenber
 * Last modified: Sat Apr 27 09:12:05 2024 mstenber
 * Edit time:     2 min
 *
 */

package cm

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"time"
)

func ToCookie(state any) (*http.Cookie, error) {
	data, err := json.Marshal(state)
	if err != nil {
		return nil, err
	}
	// TODO refresh handling
	expiration := time.Now().Add(7 * 24 * time.Hour)
	name, err := cookieName(state)
	if err != nil {
		return nil, err
	}
	value := base64.StdEncoding.EncodeToString(data)
	cookie := http.Cookie{Name: name, Value: value, Expires: expiration, Path: "/"}
	return &cookie, nil
}

func Write(w http.ResponseWriter, state any) error {
	cookie, err := ToCookie(state)
	if err != nil {
		return err
	}
	http.SetCookie(w, cookie)
	return nil
}

func ToURLValues(v url.Values, state any) (err error) {
	// Validate 'state' is what we want; we don't really need the cookie name here
	_, err = cookieName(state)
	if err != nil {
		return
	}

	// Struct we're filling from the form
	s := reflect.ValueOf(state).Elem()

	for _, field := range reflect.VisibleFields(s.Type()) {
		formKey := field.Tag.Get("cm")
		if formKey == "" {
			continue
		}

		f := s.FieldByName(field.Name)

		kind := field.Type.Kind()
		var stringified string

		switch {
		case kind == reflect.String:
			stringified = f.String()
		case kind == reflect.Bool:
			if f.Bool() {
				stringified = "true"
			} else {
				stringified = "false"
			}
		case f.CanInt():
			stringified = strconv.FormatInt(f.Int(), 10)
		case f.CanUint():
			stringified = strconv.FormatUint(f.Uint(), 10)
		default:
			err = fmt.Errorf("insupported field %s kind: %s", field.Name, kind)
		}
		if err != nil {
			return
		}
		v.Set(formKey, stringified)
	}
	return
}

func ToQueryString(base string, state any) (ret string, err error) {
	v := url.Values{}
	err = ToURLValues(v, state)
	if err != nil {
		return
	}
	ret = base
	if len(v) > 0 {
		ret = base + "?" + v.Encode()
	}
	return
}
