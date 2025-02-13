# Lixie  #

[![codecov](https://codecov.io/gh/fingon/lixie/graph/badge.svg?token=19T4DQWNFP)](https://codecov.io/gh/fingon/lixie)
[![Go Report Card](https://goreportcard.com/badge/fingon/lixie)](https://goreportcard.com/report/fingon/lixie)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

Lixie is infrastructure awareness tool. The name does not really matter and
it also does not mean anything (I am not going to use 'dog ate my homework'
excuse but instead use the modern excuse - Mistral came up with it).

The goal of Lixie is to be single binary, zero external dependency
tool. Currently there are two implemented modes:

- configuration generator for tools ( mainly [vector](https://vector.dev) )

- standalone web UI for viewing and categorizing logs ( and other events of
  interest later on )

# Log filtering - high level idea

Most of the modern logs are, to put it bluntly, spam. Typical SIEM
approaches rely on matching specific log rules of interest, and
ignoring the rest. However, this leaves the 'unknown unknown' problem
on the table, as if you do not have a rule for it, you are not aware
of it happening either (and how often it happens).

Lixie's approach is inverse - everything is worth looking at, at least
once. Lixie provides a way of cultivating ruleset about which lines are
interesting, which are not, and showing the logs based on the produced
filtering criteria. If integrated with Vector, it can also enrich the logs
with the filtering results, and subsequently Vector pipeline can e.g. drop
spam, or redirect it to something with shorter retention period.

# Functionality

- Lixie can pull logs from Loki, categorize them, and show them.

- Lixie has rule editor (and human readable dump format) for the log
  classification rules

# Demo

[Here is an example](http://www.iki.fi/fingon/lixie/). Note that only
toplevel links work - anything beyond that gets redirected to your own
local Lixie instance which probably does not exist.

# Planned TODO (There is no guarantee these actually get implemented)

## Features

- add error messages if log retrieval fails

- add better overlapping rule detection

  - with large number of rules, it is really hard to see where things
    are (and I apparently reversed the rule priority at some point
    without realizing it or perhaps I am just failing at understanding
    it now)

## Big features

- Rethink how rules are stored; just big json file can get bit unwieldy?

## Robustness

- Implement more unit tests

- Properly vendor static resources (bootstrap CSS, htmx JS)

## Prettiness

- favicon

- style the pages properly (right now just basic Bootstrap and one or two
  ugly bits remain)

## Perhaps some day?

- support victoria logs / opensearch /  quickwit as log source
