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

// When adding rules, these stream keys are NOT included
// (This doesn't prevent their manual addition)
var ignoredStreamKeys = []string{"forwarder", "host", "source_source", "source_type"}

const lokiServer = "http://fw.lan:3100"

// Loki doesn't allow empty selector; configure something which
// matches logs you want to handle with Lixie
const lokiSelector = "{forwarder=\"vector\"}"
