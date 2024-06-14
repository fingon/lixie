/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 * Created:       Mon Jun  3 07:40:40 2024 mstenber
 * Last modified: Fri Jun 14 12:30:40 2024 mstenber
 * Edit time:     18 min
 *
 */

package data

import (
	"os"
	"testing"

	"gotest.tools/v3/assert"
)

func TestDatabase(t *testing.T) {
	// Data source - static one, with exactly two log entries
	log := Log{}
	log2 := Log{}
	arr := ArraySource{Data: []*Log{&log, &log2}, Chunk: 1}

	path := "test_db.json"
	_ = os.Remove(path)

	db := Database{Path: path, Source: &arr}
	err := db.Load()
	assert.Assert(t, err != nil)

	// Add rule
	err = db.Add(LogRule{})
	assert.Equal(t, err, nil)
	assert.Equal(t, db.nextLogRuleID(), 2)

	// Ensure the logs are available
	logs, err := db.Logs()
	assert.Equal(t, len(logs), 1)
	assert.Equal(t, err, nil)

	// Ensure they also match (only one fetched now)
	assert.Equal(t, db.RuleCount(1), 1)

	// Fetch more
	logs, err = db.Logs()
	assert.Equal(t, len(logs), 2)
	assert.Equal(t, err, nil)
	assert.Equal(t, db.RuleCount(1), 2)

	// Source is empty; ensure we're still ok
	logs, err = db.Logs()
	assert.Equal(t, len(logs), 2)
	assert.Equal(t, err, nil)

	// Add another rule (using add-or-update API)
	err = db.AddOrUpdate(LogRule{})
	assert.Equal(t, err, nil)
	assert.Equal(t, len(db.LogRules.Rules), 2)

	// Update rule
	assert.Equal(t, db.LogRules.Rules[0].Ham, false)
	err = db.AddOrUpdate(LogRule{ID: 1, Ham: true})
	assert.Equal(t, err, nil)
	assert.Equal(t, db.LogRules.Rules[0].Ham, true)
	assert.Equal(t, len(db.LogRules.Rules), 2)

	// Ensure save + load gave us something similar
	db2 := Database{Path: path}
	err = db2.Load()
	assert.Equal(t, err, nil)
	assert.Equal(t, db2.LogRules.Rules[0].Ham, true)
	assert.Equal(t, len(db2.LogRules.Rules), 2)

	// Delete rule
	err = db.Delete(1)
	assert.Equal(t, err, nil)
	assert.Equal(t, len(db.LogRules.Rules), 1)
	assert.Equal(t, db.LogRules.Rules[0].Ham, false)

	assert.Equal(t, db.Delete(1), ErrRuleNotFound)
}
