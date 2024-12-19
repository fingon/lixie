/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package data

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

type LokiQueryResultDataResult struct {
	Metric map[string]string `json:"metric"`
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

type LokiQueryResultData struct {
	ResultType string                      `json:"resultType"`
	Result     []LokiQueryResultDataResult `json:"result"`
}

type LokiQueryResult struct {
	Status string               `json:"status"`
	Data   *LokiQueryResultData `json:"data"`
}

type LokiSource struct {
	Server   string
	Selector string

	lastErrorTime time.Time
	last          *Log
}

func (self *LokiSource) loadAfter(start int64) ([]*Log, error) {
	logs := []*Log{}

	base := self.Server + "/loki/api/v1/query_range"
	v := url.Values{}
	// v.Set("direction", "backward")
	v.Set("limit", "5000")
	if start > 0 {
		v.Set("query", self.Selector)
		v.Set("start", strconv.FormatInt(start, 10))
	} else {
		v.Set("query", strings.TrimSuffix(self.Selector, "}")+`,lixie!="spam"}`)
	}

	resp, err := http.Get(base + "?" + v.Encode())
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("Invalid result from Loki - status code %d", resp.StatusCode)
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
			timestamp, err := strconv.ParseInt(value[0], 10, 64)
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

func (self *LokiSource) Load() ([]*Log, error) {
	var start int64
	if self.last != nil {
		// TODO would be better to get same timestamp + eliminate if it is same entry
		start = self.last.Timestamp + 1
	}

	now := time.Now()

	if now.Sub(self.lastErrorTime).Seconds() < 5 {
		return nil, errors.New("Too short time since last failure")
	}

	logs, err := self.loadAfter(start)
	if err != nil {
		slog.Error("Loading from Loki failed", "err", err)
		self.lastErrorTime = now
		return nil, err
	}
	for i, log := range logs {
		if log.Timestamp <= start {
			logs = logs[:i]
			break
		}
	}
	if len(logs) > 0 {
		self.last = logs[0]
	}
	return logs, nil
}
