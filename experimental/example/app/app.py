import os
import socket
import sys
import requests
import random
import threading
import logging

from concurrent.futures import ThreadPoolExecutor
from flask import Flask
from redis import Redis
from werkzeug.serving import WSGIRequestHandler

app = Flask(__name__)
redis = Redis(host='redis', port=6379)
pool = ThreadPoolExecutor(max_workers=10)
sessions = threading.local()

searchapps = ['http://searchapp:8080/']

def do_redis():
  redis.incr('hits')
  return redis.get('hits')

def do_qotd():
  s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
  try:
    s.connect(("qotd", 4446))
    s.send("Hello")
    return s.recv(1024)
  finally:
    s.close()

def do_search():
  if getattr(sessions, 'session', None) == None:
    sessions.session = requests.Session()
  r = sessions.session.get(random.choice(searchapps))
  return r.text

def ignore_error(f):
  try:
    return str(f())
  except:
    logging.error("Error executing function", exc_info=sys.exc_info())
  return "Error"

@app.route('/')
def hello():
  qotd_msg = do_qotd()
  return qotd_msg

if __name__ == "__main__":
  logging.basicConfig(format='%(asctime)s %(levelname)s %(filename)s:%(lineno)d - %(message)s', level=logging.INFO)
  WSGIRequestHandler.protocol_version = "HTTP/1.1"
  app.run(host="0.0.0.0", port=80, debug=True)
