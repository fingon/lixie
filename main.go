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

	"github.com/a-h/templ"
)

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
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
