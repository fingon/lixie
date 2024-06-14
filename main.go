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

	"github.com/fingon/lixie/cm"
	"github.com/fingon/lixie/data"
)

// This is set from Makefile, which is .. ugly, and somewhat awkward
var ldBuildTimestamp = "not set"

var boot = time.Now()

// Note: While we don't have any static, double comment = static/ will be empty
// //go:embed all:static
var embedContent embed.FS

type mainConfig struct {
	RSSort int `cm:"rss"`
}

func mainHandler(st State) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config := mainConfig{RSSort: rsHits.ID()}
		err := cm.Run(r, w, &config)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		err = MainPage(st, config).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 400)
		}
	})
}

func newMux(st State) http.Handler {
	mux := http.NewServeMux()

	// Configure the routes
	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// Other options: StatusMovedPermanently, StatusFound
	//	http.Redirect(w, r, "/rule/edit", http.StatusSeeOther)
	// })
	mux.HandleFunc("/", http.NotFound)
	mux.Handle("/{$}", mainHandler(st))

	mux.Handle(topLevelLog.PathMatcher(), logListHandler(st))
	mux.Handle(topLevelLog.Path+"/{hash}/ham", logClassifyHandler(st, true))
	mux.Handle(topLevelLog.Path+"/{hash}/spam", logClassifyHandler(st, false))

	mux.Handle(topLevelLogRule.PathMatcher(), logRuleListHandler(st))
	mux.Handle(topLevelLogRule.Path+"/edit", logRuleEditHandler(st))
	mux.Handle(topLevelLogRule.Path+"/{id}/delete", logRuleDeleteSpecificHandler(st))
	mux.Handle(topLevelLogRule.Path+"/{id}/edit", logRuleEditSpecificHandler(st))
	mux.Handle("/version", versionHandler(st))

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
	args []string,
) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	// CLI
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	address := flags.String("address", "127.0.0.1", "Address to listen at")
	lokiServer := flags.String("loki-server", "https://fw.fingon.iki.fi:3100", "Address of the Loki server")
	lokiSelector := flags.String("loki-selector", "{host=~\".+\"}", "Selector to use when querying logs from Loki")
	arrayFile := flags.String("log-source-file", "", "Log file source")
	dbPath := flags.String("db", "db.json", "Database to use")
	dev := flags.Bool("dev", false, "Enable development mode")

	port := flags.Int("port", 8080, "Port number to listen at")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	var source data.LogSource
	if *arrayFile != "" {
		arr := data.ArraySource{}
		err := data.UnmarshalJSONFromPath(&arr, *arrayFile)
		if err != nil {
			return err
		}
		source = &arr
	} else {
		source = &data.LokiSource{Server: *lokiServer, Selector: *lokiSelector}
	}
	db := data.Database{Source: source, Path: *dbPath}
	err := db.Load()
	if err != nil {
		return err
	}

	state := State{DB: &db, BuildTimestamp: ldBuildTimestamp}
	if *dev {
		state.RefreshIntervalMs = 1000
	}

	mux := newMux(state)

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
