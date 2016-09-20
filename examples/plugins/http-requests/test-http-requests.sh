#!/bin/sh -e

TARGET_IP="127.0.0.1"
TARGET_PORT="8080"

# This ensures that there is always at least one connection visible in the httpd view.
# This is a workaround until https://github.com/weaveworks/scope/issues/1257 is fixed.
/bin/nc $TARGET_IP $TARGET_PORT &

while true; do
    curl http://$TARGET_IP:$TARGET_PORT &> /dev/null
    sleep 0.042
done
