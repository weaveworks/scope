import os
import socket
import sys
import random
import time
import threading
import logging


from flask import Flask
from flask import request
from werkzeug.serving import WSGIRequestHandler

app = Flask(__name__)

@app.route('/')
def hello():
  if random.random() > 0.6:
    time.sleep(2)
  else:
    time.sleep(0.5)
  return request.data

if __name__ == "__main__":
  logging.basicConfig(format='%(asctime)s %(levelname)s %(filename)s:%(lineno)d - %(message)s', level=logging.INFO)
  WSGIRequestHandler.protocol_version = "HTTP/1.0"
  app.run(host="0.0.0.0", port=80, debug=True)
