# Lixie  #

Lixie is infrastructure awareness tool. The name does not really matter and
it also does not mean anything (I am not going to use 'dog ate my homework'
excuse but instead use the modern excuse - Mistral came up with it).

The goal of Lixie is to be single binary, zero external dependency
tool. Currently there are two planned modes:

- configuration generator for tools ( mainly
  [vector](https://vector.dev) )

- standalone web UI for viewing and categorizing logs ( and other events of
  interest later on )

# Log filtering - high level idea

Most of the modern logs are, to put it bluntly, spam. Typical SIEM
approaches rely on matching specific log rules of interest, and ignoring
the rest. However, this leaves the 'unknown unknown' problem on the table,
as if you don't have rule for it, you are not aware of it happening either.

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

[Here is an example](http://www.iki.fi/fingon/lixie/). Note that only toplevel links work -
anything beyond that gets redirected to your own local Lixie instance which
probably does not exist.

To reproduce it, `wget -r -np -k -l 1 http://localhost:8080/` can be used
to get 'demo content' out of running dev instance. Deeper link depth is not
recommended as that will include edit-delete-etc link calls.

# Planned TODO (There is no guarantee these actually get implemented)

## Features

- add error messages if log retrieval fails

- add visibility for how much each rule is used in the rule editing UI

- (fts) search (for logs, rules)

## Big features

- Optimize rule evaluation (can be also used when generating rules for
  Vector); identify the stuff with most cardinality, match by that first
  (and order the rules based on that too; it doesn't hurt to make it user
  visible)

- Rethink how rules are stored; just big json file can get bit unwieldy?

- Vector configuration to add labeling based on verdict


## Robustness

- Implement more unit tests

- Properly vendor static resources (bootstrap CSS, htmx JS)

## Prettiness

- favicon

- style the pages properly (right now just basic Bootstrap and one or two
  ugly bits remain)
