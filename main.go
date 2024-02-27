/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strconv"
)

type Database struct {
	// TODO: This should be probably really a map
	LogRules []LogRule
}

func (self *Database) Add(r LogRule) {
	r.Id = self.nextLogRuleId()
	self.LogRules = append(self.LogRules, r)
}

func (self *Database) nextLogRuleId() int {
	next := 1
	for _, v := range self.LogRules {
		if v.Id >= next {
			next = v.Id + 1
		}
	}
	return next
}

type LogRuleEditHandler struct {
	Database *Database
}

func debugRequest(r *http.Request) {
	// Loop over header names
	for name, values := range r.Header {
		// Loop over all values for the name.
		for _, value := range values {
			fmt.Printf("Header %s=%s\n", name, value)
		}
	}

	for name, value := range r.Form {
		fmt.Printf("Form %s=%s\n", name, value)
	}
	fmt.Printf("\n")

}

func (self LogRuleEditHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rule, err := NewLogRuleFromForm(r)
	if err != nil {
		// TODO log error?
		return
	}
	if r.FormValue(actionSave) != "" {
		// Look for existing rule first
		for _, v := range self.Database.LogRules {
			if v.Id == rule.Id {
				v.Matchers = rule.Matchers
				http.Redirect(w, r, "/rule/", http.StatusSeeOther)
				return
			}
		}

		// Not found. Add new one.
		self.Database.Add(*rule)
		http.Redirect(w, r, "/rule/", http.StatusSeeOther)

	}
	LogRuleEdit(*rule).Render(r.Context(), w)
}

type LogRuleEditSpecificHandler struct {
	Database *Database
}

func (self LogRuleEditSpecificHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rid_string := r.PathValue("id")
	rid, err := strconv.Atoi(rid_string)
	if err != nil {
		// TODO handle error
		return
	}
	for _, v := range self.Database.LogRules {
		if v.Id == rid {
			LogRuleEdit(v).Render(r.Context(), w)
			return
		}
	}
	http.NotFound(w, r)
}

type LogRuleListHandler struct {
	Database *Database
}

func (self LogRuleListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	LogRuleList(self.Database.LogRules).Render(r.Context(), w)
}

//go:embed all:static
var embedContent embed.FS

func main() {
	// Sample content
	rules := []LogRule{
		{Id: 1,
			Matchers: []LogFieldMatcher{
				{Field: "message",
					Op:    "=",
					Value: "foobar"}}}}
	db := Database{LogRules: rules}

	static_fs, err := fs.Sub(embedContent, "static")
	if err != nil {
		log.Panic(err)
	}

	// Configure the routes
	//http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// Other options: StatusMovedPermanently, StatusFound
	//	http.Redirect(w, r, "/rule/edit", http.StatusSeeOther)
	//})
	http.Handle("/rule/", LogRuleListHandler{&db})
	http.Handle("/rule/edit", LogRuleEditHandler{&db})
	http.Handle("/rule/edit/{id}", LogRuleEditSpecificHandler{&db})
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(static_fs))))

	// CLI
	address := flag.String("address", "127.0.0.1", "Address to listen at")
	port := flag.Int("port", 8080, "Port number to listen at")

	// Start the actual server
	endpoint := fmt.Sprintf("%s:%d", *address, *port)
	fmt.Printf("Listening on %s\n", endpoint)
	http.ListenAndServe(endpoint, nil)
}
