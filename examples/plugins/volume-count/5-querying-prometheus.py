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

PLUGIN_ID="prometheus"
PLUGIN_UNIX_SOCK = "/var/run/scope/plugins/" + PLUGIN_ID + ".sock"
METRIC_NAME="http_requests_per_second"
METRIC_LABEL="HTTP req/sec"
METRIC_CONTAINER_ID_KEY="container_id"
PROMETHEUS_ADDR="prometheus.monitoring.svc.cluster.local"

def metrics():
    r = urllib2.urlopen("http://%s/api/v1/query?query=%s" % (PROMETHEUS_ADDR, METRIC_NAME))
    return json.loads(r.content).get("data", default={}).get("result", default=[])

class Handler(BaseHTTPServer.BaseHTTPRequestHandler):
    def do_GET(self):
        # The logger requires a client_address, but unix sockets don't have
        # one, so we fake it.
        self.client_address = "-"

        # Fetch and convert data from prometheus
        nodes = {}
        for metric in metrics():
            container_id = metric.get("metric", default={}).get(METRIC_CONTAINER_ID_KEY, default=None)
            if container_id == None:
                continue
            nodes["%s;<container>" % (container_id)] = {
                'metrics': {
                    METRIC_NAME: {
                        'samples': [{
                            'date': datetime.datetime.utcfromtimestamp(metric["value"][0]).isoformat('T') + 'Z',
                            'value': float(metric["value"][1]),
                        }]
                    }
                }
            }

        # Generate our json body
        body = json.dumps({
            'Plugins': [
                {
                    'id': PLUGIN_ID,
                    'label': 'Prometheus data translator',
                    'description': 'Takes data from prometheus and puts it into scope',
                    'interfaces': ['reporter'],
                    'api_version': '1',
                }
            ],
            'Container': {
                'nodes': nodes,
                'metric_templates': {
                    METRIC_NAME: {
                        'id':       METRIC_NAME,
                        'label':    METRIC_LABEL,
                        'priority': 0.1,
                    }
                }
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
