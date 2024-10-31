/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

/* Note the default locking strategy is: public functions get the
lock, private ones assume lock is taken. The exceptions have *Unlocked
suffix. */

package data

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"slices"
	"sync"

	"github.com/sourcegraph/conc/iter"
)

type LogSource interface {
	Load() ([]*Log, error)
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
	// This mutex guards log rules and logs; configuration is assumed to be static
	sync.Mutex

	// Following are essentially configuration

	// Where is this file saved
	Path   string    `json:"-"`
	Source LogSource `json:"-"`

	LogRules LogRules
	logs     []*Log

	// next id to be added state for rules
	nextID int
}

var (
	ErrHashNotFound = errors.New("specified hash not found")
	ErrRuleNotFound = errors.New("specified rule not found")
	ErrNoSource     = errors.New("source has not been specified")
)

func NewLogRules(rules []*LogRule, version int) LogRules {
	count := len(rules)
	reversed := make([]*LogRule, count)
	for k, v := range rules {
		reversed[count-k-1] = v
	}
	return LogRules{Rules: rules, Reversed: reversed, Version: version}
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

func (self *Database) updateLogs() error {
	if self.Source == nil {
		return ErrNoSource
	}
	logs, err := self.Source.Load()
	if err != nil {
		return err
	}
	self.addLogsToCounts(logs)
	self.logs = append(logs, self.logs...)
	return nil
}

func (self *Database) Logs() ([]*Log, error) {
	self.Lock()
	defer self.Unlock()

	err := self.updateLogs()
	if err != nil {
		return nil, err
	}
	return self.logs, nil
}

func (self *Database) LogCount() int {
	logs, err := self.Logs()
	if err != nil {
		return -1
	}
	return len(logs)
}

func (self *Database) getLogByHashUnlocked(hash uint64) *Log {
	self.Lock()
	defer self.Unlock()

	for _, log := range self.logs {
		if log.Hash() == hash {
			return log
		}
	}
	return nil
}

func (self *Database) ClassifyHash(hash uint64, ham bool) error {
	l := self.getLogByHashUnlocked(hash)
	if l == nil {
		return ErrHashNotFound
	}
	// Add filters - message and then all Loki labels with cardinality > 1
	rule := LogRule{Ham: ham, Matchers: []LogFieldMatcher{{
		Field: "message",
		Op:    "=",
		Value: l.Message,
	}}}
	for _, k := range l.StreamKeys {
		if ignoredStreamKeys[k] {
			continue
		}
		rule.Matchers = append(rule.Matchers, LogFieldMatcher{
			Field: k,
			Op:    "=",
			Value: l.Stream[k],
		})
	}
	return self.Add(rule)
}

func (self *Database) save(rules []*LogRule) error {
	self.LogRules = NewLogRules(rules, self.LogRules.Version+1)
	b, err := json.Marshal(self)
	if err != nil {
		return err
	}

	temp := self.Path + ".tmp"
	f, err := os.OpenFile(temp, os.O_CREATE|os.O_WRONLY, 0o644)
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
	err = os.Rename(temp, self.Path)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}

func (self *Database) addLogsToCounts(logs []*Log) {
	lrules := &self.LogRules
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
		err := self.updateLogs()
		if err != nil {
			return -1
		}
	}

	lrules := &self.LogRules

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

func (self *Database) Load() error {
	err := UnmarshalJSONFromPath(self, self.Path)
	if err != nil {
		return err
	}

	// Recreate to have also reverse slice
	self.LogRules = NewLogRules(self.LogRules.Rules, self.LogRules.Version)
	return nil
}
