#!/bin/sh
#
# Build a static Haskell binary using stack.

set -eu

if [ -z "${SRC_PATH:-}" ]; then
    echo "Must set \$SRC_PATH."
    exit 1
fi

make -C "$SRC_PATH" BUILD_IN_CONTAINER=false "$@"
