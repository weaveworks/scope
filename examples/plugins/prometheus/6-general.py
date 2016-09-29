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
QUERIES=[
    {
        'id': "http_requests_per_second",
        'label': "HTTP req/sec",
        'query': "http_requests_per_second",
        'container_id': "container_id",
        'priority': 0.1,
    },
]
PROMETHEUS_ADDR="prometheus.monitoring.svc.cluster.local"

def metrics(query):
    r = urllib2.urlopen("http://%s/api/v1/query?query=%s" % (PROMETHEUS_ADDR, query))
    return json.loads(r.content).get("data", default={}).get("result", default=[])

class Handler(BaseHTTPServer.BaseHTTPRequestHandler):
    def do_GET(self):
        # The logger requires a client_address, but unix sockets don't have
        # one, so we fake it.
        self.client_address = "-"

        # Fetch and convert data from prometheus
        nodes = {}
        for query in QUERIES:
            for metric in metrics(query):
                container_id = metric.get("metric", default={}).get(query['container_id'], default=None)
                if container_id == None:
                    continue
                nodes["%s;<container>" % (container_id)] = {
                    'metrics': {
                        query['id']: {
                            'samples': [{
                                'date': datetime.datetime.utcfromtimestamp(metric["value"][0]).isoformat('T') + 'Z',
                                'value': float(metric["value"][1]),
                            }]
                        }
                    }
                }

        # Generate our templates
        metric_templates = {}
        for i, query in QUERIES:
            metric_templates[query['id']] = query

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
                'metric_templates': metric_templates,
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
