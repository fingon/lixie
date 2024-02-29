/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/a-h/templ"
	"github.com/cespare/xxhash"
)

type Log struct {
	Timestamp  int
	Stream     map[string]string
	StreamKeys []string
	Fields     map[string]interface{}
	FieldsKeys []string
	Message    string
	RawMessage string

	// Cache the xxhash of rawMessage
	hash *uint64
}

func logLink(log *Log, op string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/log/%d/%s", log.Hash(), op))
}

func toJson(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

func (self *Log) Hash() uint64 {
	if self.hash == nil {
		hash := xxhash.Sum64([]byte(self.RawMessage))
		self.hash = &hash
	}
	return *self.hash
}

func (self *Log) IdString() string {
	return "log-" + strconv.FormatUint(self.Hash(), 10)
}

func NewLog(timestamp int, stream map[string]string, data string) *Log {
	result := Log{Timestamp: timestamp,
		Stream:     stream,
		StreamKeys: sortedKeys[string](stream),
		Message:    data,
		RawMessage: data}

	var fields map[string]interface{}
	err := json.Unmarshal([]byte(data), &fields)
	if err == nil {
		message, ok := fields["message"].(string)
		if ok {
			// Remvoe the "message" and any keys of stream from the map
			delete(fields, "message")
			for _, k := range result.StreamKeys {
				delete(fields, k)
			}
			result.Message = message
			result.Fields = fields
			result.FieldsKeys = sortedKeys[string](fields)
		}
	}
	return &result
}
