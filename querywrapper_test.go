/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 * Created:       Sun Apr 28 10:08:33 2024 mstenber
 * Last modified: Sun Apr 28 10:20:02 2024 mstenber
 * Edit time:     5 min
 *
 */

package main

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestQueryWrapper(t *testing.T) {
	qw := QueryWrapper{Base: "x"}
	qw2 := qw.Add("foo", "bar")

	// This is just chaining convenience. Ensure that is the case
	assert.Equal(t, &qw, qw2)

	assert.DeepEqual(t, qw.Values["foo"], []string{"bar"})
}
