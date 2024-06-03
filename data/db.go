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

	"github.com/sourcegraph/conc/iter"
)

type DatabaseConfig struct {
	LokiServer   string
	LokiSelector string
}

type LogRules struct {
	Rules   []*LogRule
	Version int

	// Reversed rules - these are always available if Rules are
	Reversed []*LogRule `json:"-"`

	// Internal tracking of rule matches to log lines
	rid2Count map[int]int
}

type Database struct {
	// Config is assumed to be immutable; rules and logs are not
	sync.Mutex

	config DatabaseConfig

	LogRules *LogRules
	logs     []*Log

	// next id to be added state for rules
	nextID int

	// Where is this file saved
	path string
}

var ErrHashNotFound = errors.New("specified hash not found")
var ErrRuleNotFound = errors.New("specified rule not found")

func NewLogRules(rules []*LogRule, version int) *LogRules {
	count := len(rules)
	reversed := make([]*LogRule, count)
	for k, v := range rules {
		reversed[count-k-1] = v
	}
	return &LogRules{Rules: rules, Reversed: reversed, Version: version}
}

func (self *Database) add(r LogRule) error {
	r.ID = self.nextLogRuleID()
	return self.save(append(slices.Clone(self.LogRules.Rules), &r))
}

func (self *Database) Add(r LogRule) error {
	self.Lock()
	defer self.Unlock()

	return self.add(r)
}

func (self *Database) AddOrUpdate(rule LogRule) error {
	self.Lock()
	defer self.Unlock()

	for i, v := range self.LogRules.Rules {
		if v.ID != rule.ID {
			continue
		}
		// TODO do we want to error if version differs?
		if v.Version == rule.Version {
			rule.Version++
			nrules := slices.Clone(self.LogRules.Rules)
			nrules[i] = &rule
			return self.save(nrules)
		}
		return nil
	}
	// Not found
	return self.add(rule)
}

func (self *Database) Delete(rid int) error {
	self.Lock()
	defer self.Unlock()

	for i, v := range self.LogRules.Rules {
		if v.ID == rid {
			return self.save(slices.Delete(slices.Clone(self.LogRules.Rules), i, i+1))
		}
	}
	return ErrRuleNotFound
}

func (self *Database) nextLogRuleID() int {
	id := self.nextID
	if id == 0 {
		id = 1 // Start at 1 even with empty database
		for _, v := range self.LogRules.Rules {
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

func (self *Database) updateLogs() {
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
	self.Lock()
	defer self.Unlock()

	self.updateLogs()
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
	self.Lock()
	defer self.Unlock()

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

func (self *Database) save(rules []*LogRule) error {
	lrules := NewLogRules(rules, self.LogRules.Version+1)
	self.LogRules = lrules
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

func (self *Database) addLogsToCounts(logs []*Log) {
	lrules := self.LogRules
	r2c := lrules.rid2Count
	if r2c == nil {
		return
	}
	for _, rule := range iter.Map(logs, func(logp **Log) *LogRule {
		return (*logp).ToRule(lrules)
	}) {
		if rule != nil {
			r2c[rule.ID]++
		}
	}
}

func (self *Database) RuleCount(rid int) int {
	self.Lock()
	defer self.Unlock()

	// Trigger logs refresh only if we have nothing in cache
	if self.logs == nil {
		self.updateLogs()
	}

	lrules := self.LogRules

	if lrules.rid2Count == nil {
		r2c := make(map[int]int, len(lrules.Rules))
		for _, rule := range lrules.Rules {
			r2c[rule.ID] = 0
		}
		lrules.rid2Count = r2c
		self.addLogsToCounts(self.logs)
	}
	return lrules.rid2Count[rid]
}

func NewDatabaseFromFile(config DatabaseConfig, path string) (db *Database, err error) {
	db = &Database{config: config, path: path, LogRules: &LogRules{}}
	f, err := os.Open(path)
	if err != nil {
		return
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, db)
	if err != nil {
		return
	}
	// Recreate to have also reverse slice
	db.LogRules = NewLogRules(db.LogRules.Rules, db.LogRules.Version)
	return
}
