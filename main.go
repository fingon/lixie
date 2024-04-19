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
	"github.com/fingon/lixie/data"
)

// Note: While we don't have any static, double comment = static/ will be empty
// //go:embed all:static
var embedContent embed.FS

func setupDatabase(config data.DatabaseConfig, path string) *data.Database {
	db, err := data.NewDatabaseFromFile(config, path)
	if err != nil {
		fmt.Printf("Unable to read %s: %s", path, err.Error())
	}
	return db
}

// These might be also useful at some point
//
//	getenv func(string) string,
//	stdin  io.Reader,
//	stdout, stderr io.Writer,
func run(
	_ context.Context,
	args []string) error {
	// This would be relevant only if we handled our own context.
	// However, http.ListenAndServe catches os.Interrupt so this
	// is not necessary:
	//
	// ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	// defer cancel()

	// CLI
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	address := flags.String("address", "127.0.0.1", "Address to listen at")
	lokiServer := flags.String("loki-server", "https://fw.fingon.iki.fi:3100", "Address of the Loki server")
	lokiSelector := flags.String("loki-selector", "{host=~\".+\"}", "Selector to use when querying logs from Loki")
	dbPath := flags.String("db", "db.json", "Database to use")

	port := flags.Int("port", 8080, "Port number to listen at")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	// Static content
	staticFS, err := fs.Sub(embedContent, "static")
	if err != nil {
		log.Panic(err)
	}

	config := data.DatabaseConfig{LokiServer: *lokiServer,
		LokiSelector: *lokiSelector}
	db := setupDatabase(config, *dbPath)

	// Configure the routes
	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// Other options: StatusMovedPermanently, StatusFound
	//	http.Redirect(w, r, "/rule/edit", http.StatusSeeOther)
	// })
	http.HandleFunc("/", http.NotFound)
	mainHandler := templ.Handler(MainPage())
	http.Handle("/{$}", mainHandler)

	http.Handle(topLevelLog.PathMatcher(), logListHandler(db))
	http.Handle(topLevelLog.Path+"/{hash}/ham", logClassifyHandler(db, true))
	http.Handle(topLevelLog.Path+"/{hash}/spam", logClassifyHandler(db, false))

	http.Handle(topLevelLogRule.PathMatcher(), logRuleListHandler(db))
	http.Handle(topLevelLogRule.Path+"/edit", logRuleEditHandler(db))
	http.Handle(topLevelLogRule.Path+"/{id}/delete", logRuleDeleteSpecificHandler(db))
	http.Handle(topLevelLogRule.Path+"/{id}/edit", logRuleEditSpecificHandler(db))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Start the actual server
	endpoint := fmt.Sprintf("%s:%d", *address, *port)
	fmt.Printf("Listening on %s\n", endpoint)
	return http.ListenAndServe(endpoint, nil)
}

func main() {
	ctx := context.Background()
	err := run(ctx, os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
