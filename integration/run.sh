#!/bin/sh
# Integration checks for yup-split, run inside a Debian (GNU coreutils) container.
#
# This `split` is a stream FIELD-SPLITTER (1:N line expansion), NOT GNU `split`'s
# file-chunker. GNU `split` writes input to multiple output files (xaa, xab, ...);
# this command reads lines and emits one field per output line. The two share a
# name but nothing else, so there is no GNU reference to compare against — every
# case is an `assert` against yup-split's own documented contract (see cmd-split
# COMPATIBILITY.md).
#
# assert WANT ARGS... [< STDIN]  — yup-split must produce WANT exactly.
set -eu

fails=0

assert() {
	want=$1
	shift
	got=$(yup-split "$@" 2>/dev/null || true)
	if [ "$got" = "$want" ]; then
		printf 'ok    assert  split %s\n' "$*"
	else
		printf 'FAIL  assert  split %s\n        want: %s\n        got:  %s\n' "$*" "$want" "$got"
		fails=$((fails + 1))
	fi
}

assert_stdin() {
	in=$1
	want=$2
	shift 2
	got=$(printf '%s' "$in" | yup-split "$@" 2>/dev/null || true)
	if [ "$got" = "$want" ]; then
		printf 'ok    assert  split %s < stdin\n' "$*"
	else
		printf 'FAIL  assert  split %s < stdin\n        want: %s\n        got:  %s\n' "$*" "$want" "$got"
		fails=$((fails + 1))
	fi
}

# Default: split on runs of whitespace; leading/trailing/collapsed runs are
# dropped, so adjacent spaces never yield empty fields (bytes.Fields semantics).
assert_stdin 'hello   world
  foo bar
' 'hello
world
foo
bar'

# --delimiter (-d): split on the literal delimiter, KEEPING empty fields around
# adjacent delimiters (bytes.Split semantics). "::x" -> "", "", "x".
assert_stdin 'a:b:c
::x
' 'a
b
c


x' -d :

# Multi-character delimiter is matched literally.
assert_stdin 'aXYbXYc
' 'a
b
c' -d XY

# A blank line under -d yields a single empty field (one empty output line);
# under the default whitespace split a blank line yields no fields (no output).
assert_stdin '
' '' -d :
assert_stdin '
' ''

# File operands are read in order; multiple files concatenate their fields.
printf 'one two\n' >/tmp/a.txt
printf 'three four\n' >/tmp/b.txt
assert 'one
two
three
four' /tmp/a.txt /tmp/b.txt

if [ "$fails" -ne 0 ]; then
	printf '\n%s check(s) failed\n' "$fails"
	exit 1
fi
printf '\nall checks passed\n'
