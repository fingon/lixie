/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package data

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestLogVerdictEq(t *testing.T) {
	stream := map[string]string{
		"f": "v",
	}
	log := NewLog(123, stream, `{"message": "msg", "k": "V"}`)

	verdict := LogVerdict(log, []*LogRule{})
	assert.Equal(t, verdict, LogVerdictUnknown)

	// Case: stream selector
	rule1Ok := LogRule{Matchers: []LogFieldMatcher{{Field: "f", Op: "=", Value: "v"}}}
	verdict = LogVerdict(log, []*LogRule{&rule1Ok})
	assert.Equal(t, verdict, LogVerdictSpam)

	// Case: message field
	rule2Ok := LogRule{Matchers: []LogFieldMatcher{{Field: "k", Op: "=", Value: "V"}}}
	verdict = LogVerdict(log, []*LogRule{&rule2Ok})
	assert.Equal(t, verdict, LogVerdictSpam)

	// Case: Wrong value in message
	rule3NoMatch := LogRule{Matchers: []LogFieldMatcher{{Field: "f", Op: "=", Value: "different"}}}

	verdict = LogVerdict(log, []*LogRule{&rule3NoMatch})
	assert.Equal(t, verdict, LogVerdictUnknown)

	// Case: Disabled but otherwise valid rule
	rule4Disabled := rule1Ok
	rule4Disabled.Disabled = true

	verdict = LogVerdict(log, []*LogRule{&rule4Disabled})
	assert.Equal(t, verdict, LogVerdictUnknown)

	// Case: First write wins, spam flag
	rule5Ham := rule2Ok
	rule5Ham.Ham = true
	verdict = LogVerdict(log, []*LogRule{&rule2Ok, &rule5Ham})
	assert.Equal(t, verdict, LogVerdictSpam)
}

func TestLogVerdictRe(t *testing.T) {
	stream := map[string]string{
		"f": "v",
	}
	log := NewLog(123, stream, `{"message": "msg", "k": "WHEE"}`)

	// Case: message field
	rule := LogRule{Matchers: []LogFieldMatcher{{Field: "k", Op: "=~", Value: ".*H.*"}}}
	verdict := LogVerdict(log, []*LogRule{&rule})
	assert.Equal(t, verdict, LogVerdictSpam)

	// Case: message field - ensure it is not only prefix
	rule = LogRule{Matchers: []LogFieldMatcher{{Field: "k", Op: "=~", Value: "W"}}}
	verdict = LogVerdict(log, []*LogRule{&rule})
	assert.Equal(t, verdict, LogVerdictUnknown)

	// Case: message field - ensure it is not only suffix
	rule = LogRule{Matchers: []LogFieldMatcher{{Field: "k", Op: "=~", Value: "E"}}}
	verdict = LogVerdict(log, []*LogRule{&rule})
	assert.Equal(t, verdict, LogVerdictUnknown)
}
