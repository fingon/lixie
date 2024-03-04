/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

package data

type LokiQueryResultDataResult struct {
	Metric map[string]string `json:"metric"`
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

type LokiQueryResultData struct {
	ResultType string                      `json:"resultType"`
	Result     []LokiQueryResultDataResult `json:"result"`
}

type LokiQueryResult struct {
	Status string               `json:"status"`
	Data   *LokiQueryResultData `json:"data"`
}
