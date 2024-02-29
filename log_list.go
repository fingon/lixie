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
)

type LogListConfig struct {
	AutoRefresh bool
	Expand      uint64
}

type FormValued interface {
	FormValue(string) string
}

const logListBase = "/log/"

const autoRefreshKey = "ar"
const expandKey = "exp"

func NewLogListConfig(r FormValued) LogListConfig {
	autorefresh := r.FormValue(autoRefreshKey) != ""
	expand := uint64(0)
	sv := r.FormValue(expandKey)
	if sv != "" {
		v, err := strconv.ParseUint(sv, 10, 64)
		if err == nil {
			expand = v
		}
	}
	return LogListConfig{AutoRefresh: autorefresh, Expand: expand}
}

func (self LogListConfig) WithAutoRefresh(v bool) LogListConfig {
	self.AutoRefresh = v
	return self
}

func (self LogListConfig) WithExpand(v uint64) LogListConfig {
	self.Expand = v
	return self
}

func (self LogListConfig) ToLinkString() string {
	// TODO should this be constant or parameter?
	base := logListBase
	v := url.Values{}
	if self.AutoRefresh {
		v.Set(autoRefreshKey, "1")
	}
	if self.Expand != 0 {
		v.Set(expandKey, strconv.FormatUint(self.Expand, 10))
	}
	if len(v) == 0 {
		return base
	}
	return base + "?" + v.Encode()
}

func (self LogListConfig) ToLink() templ.SafeURL {
	return templ.URL(self.ToLinkString())
}

type LogListModel struct {
	Config   LogListConfig
	Logs     []*Log
	LogRules []*LogRule
	Spam     []*Log
}

func (self *LogListModel) LogVerdict(log *Log) int {
	// TODO: These should be probably cached
	return LogVerdict(log, self.LogRules)
}

func (self *LogListModel) SplitSpam() {
	// Some spare capacity but who really cares
	logs := make([]*Log, 0, len(self.Logs))
	spam := make([]*Log, 0, len(self.Logs))
	for _, log := range self.Logs {
		if self.LogVerdict(log) != LogVerdictSpam {
			logs = append(logs, log)
		} else {
			spam = append(spam, log)
		}
	}
	self.Logs = logs
	self.Spam = spam
}

func retrieveLogs(r *http.Request) ([]*Log, error) {
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

func logListHandler(db *Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// fmt.Printf("starting logs query\n")
		logs, err := retrieveLogs(r)
		if err != nil {
			// fmt.Printf("logs query failed: %w\n", err)
			http.Error(w, err.Error(), 500)
			return
		}
		model := LogListModel{Config: NewLogListConfig(r),
			Logs:     logs,
			LogRules: db.LogRules}
		model.SplitSpam()
		LogList(model).Render(r.Context(), w)
	})
}
