#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# -*- Python -*-
#
# Author: Markus Stenberg <fingon@iki.fi>
#
# Copyright (c) 2024 Markus Stenberg
#
"""This utility script converts Lixie database of rules to Vector
( see https://vector.dev ) remap rule within Vector configuration file.

To keep evaluation fast, we split the rules by 'source' label into
binary search tree (nested set of ifs). The per-source rules could be
perhaps sorted (by common-ness of match for example), but that is left
out for now (and perhaps better left for later in any case).

Note that to keep the base vector configuration functional, probably
having the same step with nop content is the best:

e.g.

transforms:
  lixie_log:
    type: remap
    inputs:
      - remap_log
    source: |
      .lixie = "unknown"

(This code fills in the parts afterwards with matchers for ham/spam)

"""

from dataclasses import dataclass
import json


def load_rules(path):
    with open(path) as f:
        return list(reversed(json.load(f)["LogRules"]))


def get_source_matcher(rule):
    for m in rule["Matchers"]:
        if m["Field"] == "source":
            return m
    raise NotImplementedError


# Really lame, not hostile user aimed escaping.
def escape(s, outer):
    return s.replace("\\", "\\\\").replace(outer, "\\" + outer)


def split_by_source_expr(rules):
    chunk = []
    matcher_op = None
    for rule in rules:
        m = get_source_matcher(rule)
        if matcher_op != m["Op"]:
            if chunk:
                yield matcher_op, chunk
            matcher_op = m["Op"]
            chunk = []
        chunk.append(rule)
    if chunk:
        yield matcher_op, chunk


def dump_rule_matchers_ignoring_source(rule):
    prefix = ""
    for m in rule["Matchers"]:
        if (field := m["Field"]) == "source":
            continue
        op = m["Op"]
        value = m["Value"]
        if op == "=":
            evalue = escape(value, '"')
            expr = f'.{field} == "{evalue}"'
        elif op == "=~":
            evalue = escape(value, "'")
            expr = f"(parse_regex(.{field}, r'^{evalue}$') ?? null) != null"
        else:
            raise NotImplementedError
        yield f"{prefix}{expr}"
        prefix = "&& "
    if not prefix:
        yield "true"


def dump_rules_ignoring_source(chunk):
    assert chunk
    for i, rule in enumerate(chunk):
        elseprefix = "} else " if i else ""
        yield f"{elseprefix}if ("
        yield from dump_rule_matchers_ignoring_source(rule)
        yield ") {"
        verdict = "ham" if rule["Ham"] else "spam"
        yield f'.lixie = "{verdict}"'
    yield "}"


def dump_source_rules_rec(value2rules_list, stofs, endofs):
    delta = endofs - stofs
    if delta >= 4:
        ofs = stofs + delta // 2
        source = value2rules_list[ofs][0]
        yield f'if source < "{source}" ' + "{"
        yield from dump_source_rules_rec(value2rules_list, stofs, ofs)
        yield "} else {"
        yield from dump_source_rules_rec(value2rules_list, ofs, endofs)
        yield "}"
        return
    if not delta:
        return
    for ofs in range(stofs, endofs):
        source, chunk = value2rules_list[ofs]
        elseprefix = "} else " if ofs != stofs else ""
        yield f'{elseprefix}if source == "{source}" ' + "{"
        yield from dump_rules_ignoring_source(chunk)
    yield "}"


def dump_rules(rules):
    # TODO: should the field be configurable?
    yield '.lixie = "unknown"'
    yield "source = string!(.source)"
    dumped = set()
    for source_op, chunk in split_by_source_expr(rules):
        # TODO: Implement regexp support here
        assert source_op == "="
        value2rules = {}
        for rule in chunk:
            value2rules.setdefault(get_source_matcher(rule)["Value"], []).append(rule)
        value2rules_list = sorted(value2rules.items())
        yield from dump_source_rules_rec(value2rules_list, 0, len(value2rules_list))


def indent(frags):
    frags = list(frags)
    # Add newlines to the mix, where applicable
    indent = 0
    skip_next_indent = False
    for i, frag in enumerate(frags):
        nextfrag = frags[i + 1] if i < len(frags) - 1 else ""
        if frag.startswith("}"):
            indent = indent - 1
        indstring = " " * indent
        if not skip_next_indent:
            frag = indstring + frag
        if not frag.endswith("(") and not nextfrag.startswith(")"):
            frag = frag + "\n"
            skip_next_indent = False
        else:
            skip_next_indent = True
        yield frag
        # Increment indentation if necessary
        if frag.endswith("{\n"):
            indent = indent + 1


def rules_to_vrl(rules):
    lines = indent(dump_rules(rules))
    return "".join(lines)


def load_vector_config(path):
    with open(path) as f:
        if path.endswith(".yaml"):
            import yaml  # pip3 install pyyaml

            return yaml.safe_load(f)
    raise NotImplementedError


def save_vector_config(path, config):
    with open(path, "w") as f:
        if path.endswith(".yaml"):
            import yaml  # pip3 install pyyaml

            # TODO: Figure how to keep the multiline strings looking pretty (as it is, they're .. squashed..)
            yaml.dump(config, f)
            return
    raise NotImplementedError


def update_lixie_remap(config, *, name, vrl):
    transforms = config.setdefault("transforms", {})
    assert name in transforms
    transform = transforms[name]
    assert transform["type"] == "remap"
    transform["source"] = vrl


if __name__ == "__main__":
    import argparse

    p = argparse.ArgumentParser(formatter_class=argparse.ArgumentDefaultsHelpFormatter)
    p.add_argument(
        "--db",
        default="db.json",
        help="Lixie database to use",
    )
    p.add_argument(
        "--config",
        "-c",
        default="vector.yaml",
        help="Vector configuration to mutate",
    )
    p.add_argument(
        "--name", "-n", default="lixie_log", help="Vector step name to rewrite"
    )
    p.add_argument(
        "--output",
        "-o",
        required=True,
        help="Vector output configuration",
    )
    args = p.parse_args()
    rules = load_rules(args.db)
    vrl = rules_to_vrl(rules)
    config = load_vector_config(args.config)
    update_lixie_remap(config, name=args.name, vrl=vrl)
    save_vector_config(args.output, config)
