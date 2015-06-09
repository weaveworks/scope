import requests
import random
import threading
import time
import logging
import sys

app = 'http://app:5000/'
concurrency = 1

def do_requests():
  s = requests.Session()
  while True:
    try:
      s.get(app)
      logging.info("Did request")
      time.sleep(1)
    except:
      logging.error("Error doing request", exc_info=sys.exc_info())
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
