#!/usr/bin/env python

import BaseHTTPServer
import SocketServer
import datetime
import errno
import json
import os
import signal
import socket
import urllib2
from collections import namedtuple
from subprocess import check_output

PLUGIN_ID="volume-count"
PLUGIN_UNIX_SOCK="/var/run/scope/plugins/" + PLUGIN_ID + ".sock"

def run(cmd):
    return check_output(cmd).strip()

def container_volume_counts():
    # Find all containers which *should* be running "latest" of their image,
    # but there is a newer image version available
    containers = {}
    for short_id in run(["docker", "ps", "--format", "{{.ID}}"]).splitlines():
        long_id, volume_count = run(["docker", "inspect", "-f", "{{.ID}} {{.Config.Volumes | len}}", short_id]).split()
        containers[long_id.strip()] = volume_count.strip()
    return containers


class Handler(BaseHTTPServer.BaseHTTPRequestHandler):
    def do_GET(self):
        # The logger requires a client_address, but unix sockets don't have
        # one, so we fake it.
        self.client_address = "-"

        # Get current timestamp in RFC3339
        timestamp = datetime.datetime.utcnow()
        timestamp = timestamp.isoformat('T') + 'Z'

        # Fetch and convert data to scope data model
        nodes = {}
        for container_id, volume_count in container_volume_counts().iteritems():
            nodes["%s;<container>" % (container_id)] = {
                'latest': {
                    'volume_count': {
                        'timestamp': timestamp,
                        'value': volume_count,
                    }
                }
            }

        # Generate our json body
        body = json.dumps({
            'Plugins': [
                {
                    'id': PLUGIN_ID,
                    'label': 'Volume Counts',
                    'description': 'Shows how many volumes each container has mounted',
                    'interfaces': ['reporter'],
                    'api_version': '1',
                }
            ],
            'Container': {
                'nodes': nodes,
                # Templates tell the UI how to render this field.
                'metadata_templates': {
                    'volume_count': {
                        # Key where this data can be found.
                        'id': "volume_count",
                        # Human-friendly field name
                        'label': "# Volumes",
                        # Look up the 'id' in the latest object.
                        'from': "latest",
                        # Priorities over 10 are hidden, lower is earlier in the list.
                        'priority': 0.1,
                    },
                },
            },
        })

        # Send the headers
        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.send_header('Content-length', len(body))
        self.end_headers()

        # Send the body
        self.wfile.write(body)

def mkdir_p(path):
    try:
        os.makedirs(path)
    except OSError as exc:
        if exc.errno == errno.EEXIST and os.path.isdir(path):
            pass
        else:
            raise

def delete_socket_file():
    if os.path.exists(PLUGIN_UNIX_SOCK):
        os.remove(PLUGIN_UNIX_SOCK)

def sig_handler(b, a):
    delete_socket_file()
    exit(0)

def main():
    signal.signal(signal.SIGTERM, sig_handler)
    signal.signal(signal.SIGINT, sig_handler)

    mkdir_p(os.path.dirname(PLUGIN_UNIX_SOCK))
    delete_socket_file()
    server = SocketServer.UnixStreamServer(PLUGIN_UNIX_SOCK, Handler)
    try:
        server.serve_forever()
    except:
        delete_socket_file()
        raise

main()
