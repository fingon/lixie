/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package main

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"sync"
)

type Database struct {
	// TODO: There should be really locking here too;
	// log fetching is probably the more common thing though
	LogRules []*LogRule

	// Internal caching of reversed log rules (that are shown to the user)
	logRulesReversed []*LogRule

	// Log caching
	logLock sync.Mutex
	logs    []*Log

	// next id to be added state for rules
	nextId int

	// Where is this file saved
	path string
}

func (self *Database) LogRulesReversed() []*LogRule {
	if self.logRulesReversed == nil {
		count := len(self.LogRules)
		reversed := make([]*LogRule, count)
		for k, v := range self.LogRules {
			reversed[count-k-1] = v
		}
		self.logRulesReversed = reversed
	}
	return self.logRulesReversed
}

func (self *Database) Add(r LogRule) {
	r.Id = self.nextLogRuleId()
	self.LogRules = append(self.LogRules, &r)
	self.Save()
}

func (self *Database) Delete(rid int) bool {
	for i, v := range self.LogRules {
		if v.Id == rid {
			self.LogRules = slices.Delete(self.LogRules, i, i+1)
			self.Save()
			return true
		}
	}
	return false
}

func (self *Database) nextLogRuleId() int {
	id := self.nextId
	if id == 0 {
		id = 1 // Start at 1 even with empty database
		for _, v := range self.LogRules {
			if v.Id >= id {
				id = v.Id + 1
			}
		}
	}
	self.nextId = id + 1
	return id
}

func (self *Database) retrieveLogs(start int64) ([]*Log, error) {
	logs := []*Log{}

	// TBD don't hardcode endpoint and query
	base := "http://fw.lan:3100/loki/api/v1/query_range"
	v := url.Values{}
	v.Set("query", "{forwarder=\"vector\"}")
	//v.Set("direction", "backward")
	v.Set("limit", "5000")
	if start > 0 {
		v.Set("start", strconv.FormatInt(start, 10))
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

func (self *Database) Logs() []*Log {
	self.logLock.Lock()
	defer self.logLock.Unlock()

	start := int64(0)
	if len(self.logs) > 0 {
		// TODO would be better to get same timestamp + eliminate if it is same entry
		start = self.logs[0].Timestamp + 1
	}
	logs, err := self.retrieveLogs(start)
	if err != nil {
		fmt.Printf("Error retrieving logs from Loki: %s\n", err.Error())
		return self.logs
	}
	for i, log := range logs {
		if log.Timestamp <= start {
			logs = logs[:i]
			break
		}
	}
	self.logs = append(logs, self.logs...)
	return self.logs
}

func (self *Database) getLogByHash(hash uint64) *Log {
	for _, log := range self.logs {
		if log.Hash() == hash {
			return log
		}
	}
	return nil
}

func (self *Database) ClassifyHash(hash uint64, ham bool) bool {
	log := self.getLogByHash(hash)
	if log == nil {
		return false
	}
	// Add filters - message and then all Loki labels with cardinality > 1
	rule := LogRule{Ham: ham, Matchers: []LogFieldMatcher{{
		Field: "message",
		Op:    "=",
		Value: log.Message}}}
	for _, k := range log.StreamKeys {
		rule.Matchers = append(rule.Matchers, LogFieldMatcher{Field: k,
			Op:    "=",
			Value: log.Stream[k]})
	}
	self.Add(rule)
	return true
}

func (self *Database) Save() {
	self.logRulesReversed = nil

	b, err := json.Marshal(self)
	if err != nil {
		log.Panic(err)
	}

	temp := fmt.Sprintf("%s.tmp", self.path)
	f, err := os.OpenFile(temp, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Panic(err)
	}
	var ib bytes.Buffer
	json.Indent(&ib, b, "", " ")

	_, err = ib.WriteTo(f)
	if err != nil {
		log.Panic(err)
	}
	os.Rename(temp, self.path)

	defer f.Close()
}

func NewDatabaseFromFile(name string) (db *Database, err error) {
	path := "db.json"
	f, err := os.Open(path)
	data, err := io.ReadAll(f)
	if err != nil {
		return
	}
	db = &Database{}
	err = json.Unmarshal(data, db)
	db.path = name
	return
}
