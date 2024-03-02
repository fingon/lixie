/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 * Created:       Thu Feb 29 20:21:55 2024 mstenber
 * Last modified: Sat Mar  2 09:36:59 2024 mstenber
 * Edit time:     13 min
 *
 */

package main

import (
	"reflect"
	"strconv"
	"testing"

	"gotest.tools/v3/assert"
)

func logRuleToWrapper(rule *LogRule) URLWrapper {
	w := URLWrapper{}
	// TODO: Manually populate corresponding values
	// (normally browser does it so we don't have corresponding 'prod' code)
	add := func(key, value string) {
		w[key] = append(w[key], value)
	}
	add(idKey, strconv.Itoa(rule.Id))
	add(versionKey, strconv.Itoa(rule.Version))
	add(commentKey, rule.Comment)
	if rule.Ham {
		add(hamKey, "1")
	}
	if rule.Disabled {
		add(disabledKey, "1")
	}
	for i, m := range rule.Matchers {
		add(fieldId(i, fieldField), m.Field)
		add(fieldId(i, opField), m.Op)
		add(fieldId(i, valueField), m.Value)
	}
	return w
}

func TestLogRuleEndecode(t *testing.T) {
	rule := LogRule{
		Id:       42,
		Disabled: true,
		Ham:      true,
		Matchers: []LogFieldMatcher{{Field: "key",
			Op:    "=",
			Value: "value"}},
		Comment: "a comment",
		Version: 7}
	w := logRuleToWrapper(&rule)

	rule2, err := NewLogRuleFromForm(w)
	assert.Equal(t, err, nil)

	assert.Assert(t, reflect.DeepEqual(rule, *rule2))
}
