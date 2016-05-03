#!/bin/sh

set -eu

SRC=$GOPATH/src/github.com/weaveworks/go-checkpoint

# Mount the checkpoint repo:
#  -v $(pwd):/go/src/github.com/weaveworks/checkpoint

# If we run make directly, any files created on the bind mount
# will have awkward ownership.  So we switch to a user with the
# same user and group IDs as source directory.  We have to set a
# few things up so that sudo works without complaining later on.
uid=$(stat --format="%u" $SRC)
gid=$(stat --format="%g" $SRC)
echo "weave:x:$uid:$gid::$SRC:/bin/sh" >>/etc/passwd
echo "weave:*:::::::" >>/etc/shadow
echo "weave	ALL=(ALL)	NOPASSWD: ALL" >>/etc/sudoers

chmod o+rw $GOPATH/src
chmod o+rw $GOPATH/src/github.com

su weave -c "PATH=$PATH make -C $SRC BUILD_IN_CONTAINER=false $*"
