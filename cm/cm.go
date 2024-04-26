/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 * Created:       Fri Apr 26 10:35:46 2024 mstenber
 * Last modified: Fri Apr 26 21:13:04 2024 mstenber
 * Edit time:     102 min
 *
 */

package cm

import (
	"fmt"
	"net/http"
	"reflect"
)

type CookieSource interface {
	Cookie(string) (*http.Cookie, error)
}

func cookieName(state any) (string, error) {
	v := reflect.ValueOf(state)
	if v.Kind() != reflect.Pointer {
		return "", fmt.Errorf("Invalid kind: %s", v.Kind())
	}

	s := v.Elem()
	if s.Kind() != reflect.Struct {
		return "", fmt.Errorf("Invalid kind pointer: %s", s.Kind())
	}
	return fmt.Sprintf("cm-%s", s.Type()), nil
}

func GetWrapper(r *http.Request) (*URLWrapper, error) {
	if r == nil {
		return nil, nil
	}
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}
	w := URLWrapper(r.Form)
	return &w, nil
}

func RunWrapper(s CookieSource, r *URLWrapper, w http.ResponseWriter, state any) error {
	if r == nil {
		return nil
	}
	changed, err := Parse(s, r, state)
	if err != nil {
		return err
	}
	if changed {
		return Write(w, state)
	}
	return nil
}

func Run(r *http.Request, w http.ResponseWriter, state any) error {
	if r == nil {
		return nil
	}
	wr, err := GetWrapper(r)
	if err != nil {
		return err
	}
	return RunWrapper(r, wr, w, state)
}
