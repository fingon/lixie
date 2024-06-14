/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 * Created:       Sun Jun  9 20:38:22 2024 mstenber
 * Last modified: Sun Jun  9 20:42:07 2024 mstenber
 * Edit time:     3 min
 *
 */

package data

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestLogRule(t *testing.T) {
	lr := LogRule{Matchers: []LogFieldMatcher{
		{"source", "=", "dummysrc", nil},
		{"message", "=", "dummymessage", nil},
	}}
	assert.Assert(t, lr.MatchesFTS("dummysrc"))
	assert.Assert(t, lr.MatchesFTS("dummymessage"))
	assert.Assert(t, !lr.MatchesFTS("dummynonexistent"))
	assert.Equal(t, lr.SourceString(), "=dummysrc")
}
