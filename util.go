/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 * Created:       Sat Jun  1 08:54:23 2024 mstenber
 * Last modified: Sat Jun  1 08:55:00 2024 mstenber
 * Edit time:     1 min
 *
 */

package main

// TODO: These could live somewhere elsewhere too

func FilterDense[T any](slice []T, matcher func(T) bool) []T {
	n := make([]T, 0, len(slice))
	for _, e := range slice {
		if matcher(e) {
			n = append(n, e)
		}
	}
	return n
}

func FilterSparse[T any](slice []T, matcher func(T) bool) []T {
	var n []T
	for _, e := range slice {
		if matcher(e) {
			n = append(n, e)
		}
	}
	return n
}

func IsNotNil[T any](x *T) bool {
	return x != nil
}
