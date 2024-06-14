/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 * Created:       Thu May 16 07:24:25 2024 mstenber
 * Last modified: Fri Jun 14 12:12:29 2024 mstenber
 * Edit time:     32 min
 *
 */

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func retrieveURL(ctx context.Context, url string) (*http.Response, error) {
	client := http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	return client.Do(req)
}

func waitForURL(ctx context.Context, url string) error {
	for {
		resp, err := retrieveURL(ctx, url)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			fmt.Printf("Error making request: %v\n", err)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func TestMain(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	port := 18080

	go func() {
		err := run(ctx, []string{"lixie", "-port", strconv.Itoa(port), "-db", "testdata/db.json", "-log-source-file", "testdata/logs.json"})
		if err != nil {
			log.Panic(err)
		}
	}()

	ctx2, cancel2 := context.WithTimeout(ctx, 1*time.Second)
	t.Cleanup(cancel2)
	baseURL := fmt.Sprintf("http://localhost:%d", port)
	err := waitForURL(ctx2, baseURL)
	if err != nil {
		log.Panic(err)
	}

	t.Parallel()
	for _, tli := range topLevelInfos {
		t.Run(tli.Title, func(t *testing.T) {
			ctx3, cancel3 := context.WithTimeout(ctx, 1*time.Second)
			t.Cleanup(cancel3)

			resp, err := retrieveURL(ctx3, fmt.Sprintf("%s%s", baseURL, tli.Path))
			if err != nil {
				log.Panic(err)
			}
			resp.Body.Close()
			assert.Equal(t, resp.StatusCode, http.StatusOK)
		})
	}
}
