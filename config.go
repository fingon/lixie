/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package main

// TODO: These should eventually be configurable

// This is what is shown by default in log list
const primaryStreamKey = "source"

// This is global user-specific configuration from/to cookie
type GlobalConfig struct {
	// Full Text Search string
	Search string `json:"s" cm:"gsearch"`
}

const globalSearchKey = "gsearch" // must match cm: tag above
