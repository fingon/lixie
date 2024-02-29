/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package main

import (
	"cmp"
	"slices"
)

func sortedKeysWithFunc[K comparable, V any](m map[K]V, cmp func(a, b K) int) []K {
	result := make([]K, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	slices.SortFunc(result, cmp)
	return result
}

func sortedKeys[K cmp.Ordered, V any](m map[string]V) []string {
	return sortedKeysWithFunc(m, cmp.Compare)
}
