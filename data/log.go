/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package data

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/cespare/xxhash"
)

type Log struct {
	Timestamp  int64
	Time       time.Time
	Stream     map[string]string
	StreamKeys []string
	Fields     map[string]interface{}
	FieldsKeys []string
	Message    string
	RawMessage string

	// Cache the xxhash of rawMessage
	hash *uint64

	// Given ruleset version, the matching log rule (if any)
	rulesVersion int
	rule         *LogRule
}

func (self *Log) Hash() uint64 {
	if self.hash == nil {
		hash := xxhash.Sum64([]byte(self.RawMessage)) ^ uint64(self.Timestamp)
		self.hash = &hash
	}
	return *self.hash
}

func (self *Log) IDString() string {
	return "log-" + strconv.FormatUint(self.Hash(), 10)
}

func (self *Log) MatchesFTS(search string) bool {
	return strings.Contains(self.RawMessage, search)
}

func (self *Log) ToRule(rules *LogRules) *LogRule {
	if self.rulesVersion != rules.Version {
		self.rule = rules.brm.ToRule(self)
		self.rulesVersion = rules.Version
	}
	return self.rule
}

func NewLog(timestamp int64, stream map[string]string, data string) *Log {
	result := Log{
		Timestamp:  timestamp,
		Time:       time.UnixMicro(timestamp / 1000),
		Stream:     stream,
		StreamKeys: SortedKeys[string](stream),
		Message:    data,
		RawMessage: data,
	}

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
			result.FieldsKeys = SortedKeys[string](fields)
		}
	}
	return &result
}
