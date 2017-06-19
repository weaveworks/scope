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
        s.connect(("qotd.weave.local", 4446))
        s.send("Hello")
        return s.recv(1024)
    finally:
        s.close()


def do_search():
    if getattr(sessions, 'session', None) is None:
        sessions.session = requests.Session()
    r = sessions.session.get(random.choice(searchapps))
    return r.text


def do_echo(text):
    r = requests.get("http://echo.weave.local/", data=text)
    return r.text


def ignore_error(f):
    try:
        return str(f())
    except:
        logging.error("Error executing function", exc_info=sys.exc_info())
    return "Error"


# this root is for the tracing demo
@app.route('/hello')
def hello():
    qotd_msg = do_qotd()
    qotd_msg = do_echo(qotd_msg)
    return qotd_msg


# this is for normal demos
@app.route('/')
def root():
    # counter_future = pool.submit(do_redis)
    # search_future = pool.submit(do_search)
    result = do_echo(do_qotd())
    return result


if __name__ == "__main__":
    logfmt = '%(asctime)s %(levelname)s %(filename)s:%(lineno)d - %(message)s'
    logging.basicConfig(format=logfmt, level=logging.INFO)
    WSGIRequestHandler.protocol_version = "HTTP/1.1"
    app.run(host="0.0.0.0", port=80, debug=True)
