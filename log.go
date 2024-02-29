/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
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

func retrieveLogs(rules []*LogRule, r *http.Request) ([]*Log, error) {
	logs := []*Log{}

	// TBD don't hardcode endpoint and query
	base := "http://fw.lan:3100/loki/api/v1/query_range"
	v := url.Values{}
	v.Set("query", "{forwarder=\"vector\"}")

	resp, err := http.Get(base + "?" + v.Encode())
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("Invalid result from Loki: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	var result LokiQueryResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	status := result.Status
	if status != "success" {
		return nil, fmt.Errorf("invalid status from Loki:%s", status)
	}

	rtype := result.Data.ResultType
	if rtype != "streams" {
		return nil, fmt.Errorf("invalid result type from Loki:%s", rtype)
	}
	for _, result := range result.Data.Result {
		for _, value := range result.Values {
			timestamp, err := strconv.Atoi(value[0])
			if err != nil {
				return nil, err
			}
			logs = append(logs, NewLog(timestamp, result.Stream, value[1]))
		}
	}

	// Loki output is by metric/stream and then by time; we don't
	// really care, sort by timestamp desc. This may need to be
	// rethought when we fetch more than just the latest
	slices.SortFunc(logs, func(a, b *Log) int {
		return -cmp.Compare(a.Timestamp, b.Timestamp)
	})

	return logs, nil
}
