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
	rule1_ok := LogRule{Matchers: []LogFieldMatcher{{Field: "f", Op: "=", Value: "v"}}}
	verdict = LogVerdict(log, []*LogRule{&rule1_ok})
	assert.Equal(t, verdict, LogVerdictSpam)

	// Case: message field
	rule2_ok := LogRule{Matchers: []LogFieldMatcher{{Field: "k", Op: "=", Value: "V"}}}
	verdict = LogVerdict(log, []*LogRule{&rule2_ok})
	assert.Equal(t, verdict, LogVerdictSpam)

	// Case: Wrong value in message
	rule3_no_match := LogRule{Matchers: []LogFieldMatcher{{Field: "f", Op: "=", Value: "different"}}}

	verdict = LogVerdict(log, []*LogRule{&rule3_no_match})
	assert.Equal(t, verdict, LogVerdictUnknown)

	// Case: Disabled but otherwise valid rule
	rule4_disabled := rule1_ok
	rule4_disabled.Disabled = true

	verdict = LogVerdict(log, []*LogRule{&rule4_disabled})
	assert.Equal(t, verdict, LogVerdictUnknown)

	// Case: First write wins, spam flag
	rule5_ham := rule2_ok
	rule5_ham.Ham = true
	verdict = LogVerdict(log, []*LogRule{&rule2_ok, &rule5_ham})
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
