#!/usr/bin/env python

import BaseHTTPServer
import SocketServer
import errno
import os
import signal
import socket

PLUGIN_ID="prometheus"
PLUGIN_UNIX_SOCK = "/var/run/scope/plugins/" + PLUGIN_ID + ".sock"

class Handler(BaseHTTPServer.BaseHTTPRequestHandler):
    def do_GET(self):
        # The logger requires a client_address, but unix sockets don't have
        # one, so we fake it.
        self.client_address = "-"

        # Send the headers
        self.send_response(200)
        self.end_headers()

        # Send the body
        self.wfile.write(b"foo")

def mkdir_p(path):
    try:
        os.makedirs(path)
    except OSError as exc:
        if exc.errno == errno.EEXIST and os.path.isdir(path):
            pass
        else:
            raise

def main():
    mkdir_p(os.path.dirname(PLUGIN_UNIX_SOCK))
    server = SocketServer.UnixStreamServer(PLUGIN_UNIX_SOCK, Handler)
    server.serve_forever()

main()
