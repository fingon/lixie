/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package data

// TODO: These should eventually be configurable

// When adding rules, these stream keys are NOT included
//
// (This doesn't prevent their manual addition)
var ignoredStreamKeys = []string{"forwarder", "host", "source_source", "source_type"}
