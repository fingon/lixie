/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 * Created:       Sun Jun  2 19:57:04 2024 mstenber
 * Last modified: Sun Jun  2 20:27:04 2024 mstenber
 * Edit time:     3 min
 *
 */

package main

import "github.com/fingon/lixie/data"

type State struct {
	// How often we retry page load
	RefreshIntervalMs int

	// Current build version
	BuildTimestamp string

	DB *data.Database
}
