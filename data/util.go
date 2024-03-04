/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package data

import (
	"cmp"
	"slices"
)

func SortedKeysWithFunc[K comparable, V any](m map[K]V, cmp func(a, b K) int) []K {
	result := make([]K, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	slices.SortFunc(result, cmp)
	return result
}

func SortedKeys[K cmp.Ordered, V any](m map[string]V) []string {
	return SortedKeysWithFunc(m, cmp.Compare)
}
