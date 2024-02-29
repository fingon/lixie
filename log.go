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
	"sort"
	"strconv"
	"strings"

	"github.com/cespare/xxhash"
)

type Log struct {
	Timestamp int
	Stream    map[string]string
	Fields    map[string]interface{}
	Message   string
}

func (self *Log) Hash() uint64 {
	return xxhash.Sum64([]byte(self.Message))
}

func (self *Log) StreamString() string {
	parts := []string{}
	for _, k := range sortedKeys[string](self.Stream) {
		v := self.Stream[k]
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	// TODO this isn't particularly performant, we should really
	// centralize all 'same' Streams as distinct objects
	parts = sort.StringSlice(parts)
	return strings.Join(parts, ",")
}

func NewLog(timestamp int, stream map[string]string, data string) *Log {
	result := Log{Timestamp: timestamp,
		Stream:  stream,
		Message: data}
	var content map[string]interface{}
	err := json.Unmarshal([]byte(data), &content)
	if err == nil {
		message, ok := content["message"].(string)
		if ok {
			result.Message = message
			result.Fields = content
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
