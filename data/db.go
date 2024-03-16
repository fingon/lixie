/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package data

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"sync"
)

type DatabaseConfig struct {
	LokiServer   string
	LokiSelector string
}

type Database struct {
	config DatabaseConfig

	// TODO: There should be really locking here too;
	// log fetching is probably the more common thing though
	LogRules     []*LogRule
	rulesVersion int
	rid2Count    map[int]int

	// Internal caching of reversed log rules (that are shown to the user)
	logRulesReversed []*LogRule

	// Log caching
	logLock sync.Mutex
	logs    []*Log

	// next id to be added state for rules
	nextID int

	// Where is this file saved
	path string
}

var ErrHashNotFound = errors.New("specified hash not found")
var ErrRuleNotFound = errors.New("specified rule not found")

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

func (self *Database) Add(r LogRule) error {
	r.ID = self.nextLogRuleID()
	self.LogRules = append(self.LogRules, &r)
	return self.Save()
}

func (self *Database) Delete(rid int) error {
	for i, v := range self.LogRules {
		if v.ID == rid {
			self.LogRules = slices.Delete(self.LogRules, i, i+1)
			return self.Save()
		}
	}
	return ErrRuleNotFound
}

func (self *Database) nextLogRuleID() int {
	id := self.nextID
	if id == 0 {
		id = 1 // Start at 1 even with empty database
		for _, v := range self.LogRules {
			if v.ID >= id {
				id = v.ID + 1
			}
		}
	}
	self.nextID = id + 1
	return id
}

func (self *Database) retrieveLogs(start int64) ([]*Log, error) {
	logs := []*Log{}

	base := self.config.LokiServer + "/loki/api/v1/query_range"
	v := url.Values{}
	v.Set("query", self.config.LokiSelector)
	// v.Set("direction", "backward")
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

func (self *Database) updateLogsWithLock() {
	start := int64(0)
	if len(self.logs) > 0 {
		// TODO would be better to get same timestamp + eliminate if it is same entry
		start = self.logs[0].Timestamp + 1
	}
	logs, err := self.retrieveLogs(start)
	if err != nil {
		fmt.Printf("Error retrieving logs from Loki: %s\n", err.Error())
	}
	for i, log := range logs {
		if log.Timestamp <= start {
			logs = logs[:i]
			break
		}
	}
	self.addLogsToCounts(logs)
	self.logs = append(logs, self.logs...)
}

func (self *Database) Logs() []*Log {
	self.logLock.Lock()
	defer self.logLock.Unlock()
	self.updateLogsWithLock()
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

func (self *Database) ClassifyHash(hash uint64, ham bool) error {
	l := self.getLogByHash(hash)
	if l == nil {
		return ErrHashNotFound
	}
	// Add filters - message and then all Loki labels with cardinality > 1
	rule := LogRule{Ham: ham, Matchers: []LogFieldMatcher{{
		Field: "message",
		Op:    "=",
		Value: l.Message}}}
	for _, k := range l.StreamKeys {
		for _, ignoredStream := range ignoredStreamKeys {
			if ignoredStream == k {
				goto next
			}
		}
		rule.Matchers = append(rule.Matchers, LogFieldMatcher{Field: k,
			Op:    "=",
			Value: l.Stream[k]})
	next:
	}
	return self.Add(rule)
}

func (self *Database) Save() error {
	self.logRulesReversed = nil
	self.rulesVersion++
	self.rid2Count = nil

	b, err := json.Marshal(self)
	if err != nil {
		return err
	}

	temp := self.path + ".tmp"
	f, err := os.OpenFile(temp, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	var ib bytes.Buffer
	err = json.Indent(&ib, b, "", " ")
	if err != nil {
		return err
	}

	_, err = ib.WriteTo(f)
	if err != nil {
		return err
	}
	err = os.Rename(temp, self.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}

func (self *Database) LogToRule(log *Log) *LogRule {
	// TODO: Does the locking here matter?
	return log.ToRule(self.rulesVersion, self.LogRulesReversed())
}

func (self *Database) addLogsToCounts(logs []*Log) {
	r2c := self.rid2Count
	if r2c == nil {
		return
	}
	for _, log := range logs {
		rule := self.LogToRule(log)
		if rule != nil {
			r2c[rule.ID]++
		}
	}
}

func (self *Database) RuleCount(rid int) int {
	self.logLock.Lock()
	defer self.logLock.Unlock()

	// Trigger logs refresh only if we have nothing in cache
	if self.logs == nil {
		self.updateLogsWithLock()
	}

	if self.rid2Count == nil {
		r2c := make(map[int]int, len(self.LogRules))
		for _, rule := range self.LogRules {
			r2c[rule.ID] = 0
		}
		self.rid2Count = r2c
		self.addLogsToCounts(self.logs)
	}
	return self.rid2Count[rid]
}

func NewDatabaseFromFile(config DatabaseConfig, path string) (db *Database, err error) {
	db = &Database{config: config, path: path, rulesVersion: 1}
	f, err := os.Open(path)
	if err != nil {
		return
	}
	data, err := io.ReadAll(f)
	if err == nil {
		err = json.Unmarshal(data, db)
	}
	return
}
