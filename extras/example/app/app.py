import socket
import sys
import requests
import threading
import logging
import argparse

from concurrent.futures import ThreadPoolExecutor
from flask import Flask
from redis import Redis
from werkzeug.serving import WSGIRequestHandler

app = Flask(__name__)
redis = Redis(host='redis', port=6379)
pool = ThreadPoolExecutor(max_workers=10)
sessions = threading.local()
args = None


def do_redis():
    redis.incr('hits')
    return redis.get('hits')


def do_qotd():
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    try:
        s.settimeout(args.timeout)
        s.connect((args.qotd, 4446))
        s.send("Hello")
        return s.recv(1024)
    finally:
        s.close()


def do_search():
    if getattr(sessions, 'session', None) is None:
        sessions.session = requests.Session()
    r = sessions.session.get(args.search, timeout=args.timeout)
    return r.text


def do_echo(text):
    if getattr(sessions, 'session', None) is None:
        sessions.session = requests.Session()
    r = sessions.session.get(args.echo, data=text, timeout=args.timeout)
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
    counter_future = pool.submit(do_redis)
    search_future = pool.submit(do_search)
    qotd_future = pool.submit(do_qotd)
    echo_future = pool.submit(lambda: do_echo("foo"))
    result = 'Hello World! I have been seen %s times.' % ignore_error(
        counter_future.result)
    result += ignore_error(search_future.result)
    result += ignore_error(qotd_future.result)
    result += ignore_error(echo_future.result)
    return result


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument('-redis', default="redis.weave.local")
    parser.add_argument('-search', default="http://search.weave.local:80/")
    parser.add_argument('-qotd', default="qotd.weave.local")
    parser.add_argument('-echo', default="http://echo.weave.local:80/")
    parser.add_argument('-timeout', default=0.5, type=float)
    args = parser.parse_args()

    logfmt = '%(asctime)s %(levelname)s %(filename)s:%(lineno)d - %(message)s'
    logging.basicConfig(format=logfmt, level=logging.INFO)
    WSGIRequestHandler.protocol_version = "HTTP/1.1"
    app.run(host="0.0.0.0", port=80, debug=True)
