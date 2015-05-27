import requests
import random
import threading
import logging
import sys

pyapps = ['http://pyapp1:5000/', 'http://pyapp2:5000/']
concurrency = 5

def do_requests():
  s = requests.Session()
  while True:
    try:
      s.get(random.choice(pyapps))
    except:
      logging.error("Error doing request", exc_info=sys.exc_info())
    logging.info("Did request")

def main():
  logging.info("Starting %d thread", concurrency)
  threads = [threading.Thread(target=do_requests) for i in range(concurrency)]
  for thread in threads:
    thread.start()
  for thread in threads:
    thread.join()
  logging.info("Exiting")

if __name__ == "__main__":
  logging.basicConfig()
  main()
