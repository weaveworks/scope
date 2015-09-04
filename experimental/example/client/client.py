import requests
import random
import threading
import time
import logging
import socket
import sys

frontend = 'frontend'
concurrency = 2

def do_request(s):
  addrs = socket.getaddrinfo(frontend, 80)
  if len(addrs) <= 0:
    return
  addr = random.choice(addrs)
  s.get("http://%s:%d" % addr[4], timeout=1.0)

def do_requests():
  s = requests.Session()
  while True:
    try:
      do_request(s)
    except:
      logging.error("Error doing request", exc_info=sys.exc_info())

    time.sleep(1)
    logging.info("Did request")

def main():
  logging.info("Starting %d threads", concurrency)
  threads = [threading.Thread(target=do_requests) for i in range(concurrency)]
  for thread in threads:
    thread.start()
  for thread in threads:
    thread.join()
  logging.info("Exiting")

if __name__ == "__main__":
  logging.basicConfig(format='%(asctime)s %(levelname)s %(filename)s:%(lineno)d - %(message)s', level=logging.INFO)
  main()
