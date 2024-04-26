/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 * Created:       Fri Apr 26 10:35:46 2024 mstenber
 * Last modified: Fri Apr 26 14:38:05 2024 mstenber
 * Edit time:     93 min
 *
 */

package cm

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
)

type CookieSource interface {
	Cookie(string) (*http.Cookie, error)
}

var ErrNoSource = errors.New("specified cookie source is nil")

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

func Run(r *http.Request, w http.ResponseWriter, state any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	changed, err := Parse(r, URLWrapper(r.Form), state)
	if err != nil {
		return err
	}
	if changed {
		fmt.Printf("XXX changed\n")
		return Write(w, state)
	}
	return nil
}
