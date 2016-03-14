import argparse
import requests
import random
import threading
import time
import logging
import socket
import sys

def do_request(s, args):
  addrs = socket.getaddrinfo(args.target, 80)
  addrs = [a
    for a in addrs
    if a[0] == socket.AF_INET]
  logging.info("got %s", addrs)
  if len(addrs) <= 0:
    return
  addr = random.choice(addrs)
  s.get("http://%s:%d" % addr[4], timeout=1.0)

def do_requests(args):
  s = requests.Session()
  while True:
    try:
      if args.persist:
        do_request(s, args)
      else:
        do_request(requests.Session(), args)
    except:
      logging.error("Error doing request", exc_info=sys.exc_info())

    time.sleep(1)
    logging.info("Did request")

def main():
  parser = argparse.ArgumentParser()
  parser.add_argument('-target', default="frontend")
  parser.add_argument('-concurrency', default=2, type=int)
  parser.add_argument('-persist', default=True, type=bool)
  args = parser.parse_args()

  logging.info("Starting %d threads, targeting %s", args.concurrency, args.target)
  threads = [threading.Thread(target=do_requests, args=(args,))
    for i in range(args.concurrency)]
  for thread in threads:
    thread.start()
  for thread in threads:
    thread.join()
  logging.info("Exiting")

if __name__ == "__main__":
  logging.basicConfig(format='%(asctime)s %(levelname)s %(filename)s:%(lineno)d - %(message)s', level=logging.INFO)
  main()
