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

# Functionality

- Lixie can pull logs from Loki, categorize them, and show them.

# Planned TODO (There is no guarantee these actually get implemented)

## Features

- add visibility for how much each rule is used in the rule editing UI

- add visibility for matching in the rule editing UI (dump matching log
  rule list in expanded format)

- better log list selection features (current non-spam is ok default, but
  also seeing spam, ham only, all might be valid?)

- configuration file / CLI for the hardcoded stuff (e.g. Loki URL, database
  filename)

- (fts) search (for logs, rules)

- regexp match in addition to simple = (And convert the op to popup
  selection instead of text input)

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
