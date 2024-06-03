/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 * Created:       Mon Jun  3 07:40:40 2024 mstenber
 * Last modified: Mon Jun  3 08:01:57 2024 mstenber
 * Edit time:     7 min
 *
 */

package data

import (
	"os"
	"testing"

	"gotest.tools/v3/assert"
)

func TestDatabase(t *testing.T) {
	path := "test_db.json"
	_ = os.Remove(path)
	db, err := NewDatabaseFromFile(DatabaseConfig{}, path)
	assert.Assert(t, err != nil)
	// Add rule
	err = db.Add(LogRule{})
	assert.Equal(t, err, nil)

	// Add another rule
	err = db.Add(LogRule{})
	assert.Equal(t, err, nil)
	assert.Equal(t, len(db.LogRules.Rules), 2)

	// Update rule
	assert.Equal(t, db.LogRules.Rules[0].Ham, false)
	err = db.AddOrUpdate(LogRule{ID: 1, Ham: true})
	assert.Equal(t, err, nil)
	assert.Equal(t, db.LogRules.Rules[0].Ham, true)
	assert.Equal(t, len(db.LogRules.Rules), 2)

	// Ensure save + load gave us something similar
	db2, err := NewDatabaseFromFile(DatabaseConfig{}, path)
	assert.Equal(t, err, nil)
	assert.Equal(t, db2.LogRules.Rules[0].Ham, true)
	assert.Equal(t, len(db2.LogRules.Rules), 2)

	// Delete rule
	err = db.Delete(1)
	assert.Equal(t, err, nil)
	assert.Equal(t, len(db.LogRules.Rules), 1)
	assert.Equal(t, db.LogRules.Rules[0].Ham, false)
}
