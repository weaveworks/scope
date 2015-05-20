import os
import socket
import requests
from concurrent.futures import ThreadPoolExecutor
from flask import Flask
from redis import Redis

app = Flask(__name__)
redis = Redis(host='redis', port=6379)
pool = ThreadPoolExecutor(max_workers=10)

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
  r = requests.get('http://goapp:8080/')
  return r.text


@app.route('/')
def hello():
  counter_future = pool.submit(do_redis)
  search_future = pool.submit(do_search)
  qotd_future = pool.submit(do_qotd)
  result = 'Hello World! I have been seen %s times.' % counter_future.result()
  result += search_future.result()
  result += qotd_future.result()
  return result

if __name__ == "__main__":
  app.run(host="0.0.0.0", debug=True)
