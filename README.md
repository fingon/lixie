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

# TODO

## Features

- configuration file / CLI for the hardcoded stuff (e.g. Loki URL, database
  filename)

## Big features

- Vector configuration to add labeling based on verdict

## Robustness

- Implement more unit tests

- Properly vendor static resources (bootstrap CSS, htmx JS)

## Prettiness

- favicon

- style the pages properly (right now just basic Bootstrap and one or two
  ugly bits remain)
