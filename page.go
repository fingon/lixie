/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package main

import (
	"github.com/a-h/templ"
)

const (
	TopLevelMain int = iota
	TopLevelLog
	TopLevelLogRule
)

type PageInfo struct {
	ID    int
	Title string
	Path  string
}

func (self *PageInfo) PathMatcher() string {
	return self.Path + "/{$}"
}

func (self *PageInfo) URL() templ.SafeURL {
	return templ.URL(self.Path)
}

var topLevelLog = PageInfo{TopLevelLog, "Logs", "/log"}
var topLevelLogRule = PageInfo{TopLevelLogRule, "Log rules", "/log/rule"}

var topLevelInfos = []PageInfo{topLevelLog, topLevelLogRule}

var logRuleEdit = PageInfo{Path: "/log/rule/edit"}
