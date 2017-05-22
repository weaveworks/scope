#!/bin/sh

set -eu

if [ -n "${SRC_NAME:-}" ]; then
    SRC_PATH=${SRC_PATH:-$GOPATH/src/$SRC_NAME}
elif [ -z "${SRC_PATH:-}" ]; then
    echo "Must set either \$SRC_NAME or \$SRC_PATH."
    exit 1
fi

# If we run make directly, any files created on the bind mount
# will have awkward ownership.  So we switch to a user with the
# same user and group IDs as source directory.  We have to set a
# few things up so that sudo works without complaining later on.
uid=$(stat --format="%u" "$SRC_PATH")
gid=$(stat --format="%g" "$SRC_PATH")
echo "weave:x:$uid:$gid::$SRC_PATH:/bin/sh" >>/etc/passwd
echo "weave:*:::::::" >>/etc/shadow
echo "weave	ALL=(ALL)	NOPASSWD: ALL" >>/etc/sudoers

su weave -c "PATH=$PATH make -C $SRC_PATH BUILD_IN_CONTAINER=false $*"
