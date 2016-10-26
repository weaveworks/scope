#!/bin/sh

set -e

# shellcheck disable=SC2046
./in_parallel.sh "make RM=" $(find . -maxdepth 2 -name "./*.go" -printf "%h\n" | sort -u | sed -n 's/\.\/\(.*\)/\1\/\1/p')
