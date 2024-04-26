/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package main

type FTSMatchable interface {
	MatchesFTS(string) bool
}

func filterFTS[T FTSMatchable](s []T, search string, limit int) []T {
	if search == "" {
		return s
	}
	if len(s) < limit {
		limit = len(s)
	}
	filtered := make([]T, 0, limit)
	for _, o := range s {
		if o.MatchesFTS(search) {
			filtered = append(filtered, o)
		}
	}
	return filtered
}
