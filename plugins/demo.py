#!/usr/bin/python
#
# Example Scope probe plugin that annotates processes
# on the local system with details from their environment.
#
import sys
from BaseHTTPServer import BaseHTTPRequestHandler
from httplib import HTTPResponse

def parse_request():
  req = BaseHTTPRequestHandler()
  req.rfile = sys.stdin
  req.raw_requestline = req.rfile.readline()
  req.parse_request()
  return req

def send_response(data):
  data_json = json.dumps(data)
  sys.stdout.write("HTTP/1.0 200 OK\n")
  sys.stdout.write("Content-Length: %d\n" % len(data_json))
  sys.stdout.write("\n")
  sys.stdout.write(data_json)

def main():
  req = parse_request()
  if req.command == "POST" and req.path == "/tag":
    report = json.load(req.rfile)
    report = tag()
  elif req.command == "GET" and req.path == "/report":
    report = report()

if __name__ == "__main__":
  main()
