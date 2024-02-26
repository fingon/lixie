/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package main

import (
	"flag"
	"fmt"
	"net/http"
)

type Database struct {
	LogRules []LogRule
}

type LogRuleEditHandler struct {
	Database *Database
}

func (self LogRuleEditHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("ServeHTTP LogRuleEditHandler\n")

	if err := r.ParseForm(); err != nil {
		// TODO log error
		return
	}

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

	rule, err := NewLogRuleFromForm(r)
	if err != nil {
		// TODO log error?
		return
	}
	LogRuleEdit(*rule).Render(r.Context(), w)
}

func main() {
	db := Database{LogRules: []LogRule{}}

	address := flag.String("address", "127.0.0.1", "Address to listen at")
	port := flag.Int("port", 8080, "Port number to listen at")

	// TODO: Consider if this should instead use go:embed, as all
	// we really need are few icons and stylesheet
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Other options: StatusMovedPermanently, StatusFound
		http.Redirect(w, r, "/rule/edit", http.StatusSeeOther)
	})

	http.Handle("/rule/edit", LogRuleEditHandler{&db})

	endpoint := fmt.Sprintf("%s:%d", *address, *port)
	fmt.Printf("Listening on %s\n", endpoint)
	http.ListenAndServe(endpoint, nil)
}
