#!/bin/sh

set -e

./in_parallel.sh "make RM=" tracer $(find . -maxdepth 2 -name *.go -printf "%h\n" | sort -u | sed -n 's/\.\/\(.*\)/\1\/\1/p')
