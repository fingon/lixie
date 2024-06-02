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
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/a-h/templ"
	"github.com/fingon/lixie/data"
)

var boot = time.Now()

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

func newMux(db *data.Database) http.Handler {
	mux := http.NewServeMux()

	// Configure the routes
	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// Other options: StatusMovedPermanently, StatusFound
	//	http.Redirect(w, r, "/rule/edit", http.StatusSeeOther)
	// })
	mux.HandleFunc("/", http.NotFound)
	mainHandler := templ.Handler(MainPage(db))
	mux.Handle("/{$}", mainHandler)

	mux.Handle(topLevelLog.PathMatcher(), logListHandler(db))
	mux.Handle(topLevelLog.Path+"/{hash}/ham", logClassifyHandler(db, true))
	mux.Handle(topLevelLog.Path+"/{hash}/spam", logClassifyHandler(db, false))

	mux.Handle(topLevelLogRule.PathMatcher(), logRuleListHandler(db))
	mux.Handle(topLevelLogRule.Path+"/edit", logRuleEditHandler(db))
	mux.Handle(topLevelLogRule.Path+"/{id}/delete", logRuleDeleteSpecificHandler(db))
	mux.Handle(topLevelLogRule.Path+"/{id}/edit", logRuleEditSpecificHandler(db))
	mux.Handle("/version", versionHandler())

	// Static content
	staticFS, err := fs.Sub(embedContent, "static")
	if err != nil {
		log.Panic(err)
	}

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	return mux
}

// Most of this function is based on
// https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/
//
// These might be also useful at some point
//
//	getenv func(string) string,
//	stdin  io.Reader,
//	stdout, stderr io.Writer,
func run(
	ctx context.Context,
	args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

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

	config := data.DatabaseConfig{LokiServer: *lokiServer,
		LokiSelector: *lokiSelector}
	db := setupDatabase(config, *dbPath)

	mux := newMux(db)

	// Start the actual server
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(*address, strconv.Itoa(*port)),
		Handler: mux,
	}
	go func() {
		log.Printf("listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening: %v", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	}()
	wg.Wait()
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
