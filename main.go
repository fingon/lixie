/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"

	"github.com/a-h/templ"
)

type Database struct {
	// TODO: This should be probably really a map
	LogRules         []*LogRule
	logRulesReversed []*LogRule
	nextId           int
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
	self.logRulesReversed = nil
}

func (self *Database) Delete(rid int) bool {
	for i, v := range self.LogRules {
		if v.Id == rid {
			self.LogRules = slices.Delete(self.LogRules, i, i+1)
			self.logRulesReversed = nil
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

type LogRuleEditHandler struct {
}

func logRuleEditHandler(db *Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		rule, err := NewLogRuleFromForm(r)
		if err != nil {
			// TODO log error?
			return
		}
		if r.FormValue(actionSave) != "" {
			// Look for existing rule first
			for _, v := range db.LogRules {
				if v.Id == rule.Id {
					fmt.Printf("Rewrote matchers of rule %d\n", v.Id)
					v.Matchers = rule.Matchers
					http.Redirect(w, r, "/rule/", http.StatusSeeOther)
					return
				}
			}

			// Not found. Add new one.
			fmt.Printf("Adding new rule\n")
			db.Add(*rule)
			http.Redirect(w, r, "/rule/", http.StatusSeeOther)
			return
		}
		LogRuleEdit(*rule).Render(r.Context(), w)
	})
}

func logRuleDeleteSpecificHandler(db *Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid_string := r.PathValue("id")
		rid, err := strconv.Atoi(rid_string)
		if err != nil {
			// TODO handle error
			return
		}
		if db.Delete(rid) {
			http.Redirect(w, r, "/rule/", http.StatusSeeOther)
			return
		}
		http.NotFound(w, r)
	})
}

func logRuleEditSpecificHandler(db *Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid_string := r.PathValue("id")
		rid, err := strconv.Atoi(rid_string)
		if err != nil {
			// TODO handle error
			return
		}
		for _, v := range db.LogRules {
			if v.Id == rid {
				LogRuleEdit(*v).Render(r.Context(), w)
				return
			}
		}
		http.NotFound(w, r)
	})
}

func logRuleListHandler(db *Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		LogRuleList(db.LogRulesReversed()).Render(r.Context(), w)

	})
}

func logListHandler(db *Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// fmt.Printf("starting logs query\n")
		logs, err := retrieveLogs(db.LogRules, r)
		if err != nil {
			// fmt.Printf("logs query failed: %w\n", err)
			http.Error(w, err.Error(), 500)
			return
		}
		LogList(r.FormValue("autorefresh") != "", logs).Render(r.Context(), w)
	})
}

// Note: While we don't have any static, double comment = static/ will be empty
// //go:embed all:static
var embedContent embed.FS

// These might be also useful at some point
//
//	getenv func(string) string,
//	stdin  io.Reader,
//	stdout, stderr io.Writer,
func run(
	ctx context.Context,
	args []string) error {

	// This would be relevant only if we handled our own context.
	// However, http.ListenAndServe catches os.Interrupt so this
	// is not necessary:
	//
	//ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	//defer cancel()

	// Sample content
	rule := LogRule{Id: 1,
		Matchers: []LogFieldMatcher{
			{Field: "message",
				Op:    "=",
				Value: "foobar"}}}
	rules := []*LogRule{&rule}
	db := Database{LogRules: rules}

	// CLI
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	address := flags.String("address", "127.0.0.1", "Address to listen at")
	port := flags.Int("port", 8080, "Port number to listen at")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	// Static content
	static_fs, err := fs.Sub(embedContent, "static")
	if err != nil {
		log.Panic(err)
	}

	// Configure the routes
	//http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// Other options: StatusMovedPermanently, StatusFound
	//	http.Redirect(w, r, "/rule/edit", http.StatusSeeOther)
	//})
	http.HandleFunc("/", http.NotFound)
	main_handler := templ.Handler(MainPage())
	http.Handle("/{$}", main_handler)
	http.Handle("/index.html", main_handler)
	http.Handle("/log/{$}", logListHandler(&db))
	http.Handle("/rule/{$}", logRuleListHandler(&db))
	http.Handle("/rule/edit", logRuleEditHandler(&db))
	http.Handle("/rule/{id}/delete", logRuleDeleteSpecificHandler(&db))
	http.Handle("/rule/{id}/edit", logRuleEditSpecificHandler(&db))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(static_fs))))

	// Start the actual server
	endpoint := fmt.Sprintf("%s:%d", *address, *port)
	fmt.Printf("Listening on %s\n", endpoint)
	http.ListenAndServe(endpoint, nil)
	return nil
}

func main() {
	ctx := context.Background()
	err := run(ctx, os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%w\n", err)
		os.Exit(1)
	}
}
