import argparse
import requests
import random
import threading
import time
import logging
import socket
import sys


def do_request(s, args):
    addrs = socket.getaddrinfo(args.target, args.port)
    addrs = [a for a in addrs if a[0] == socket.AF_INET]
    if len(addrs) <= 0:
        logging.info("Could not resolve %s", args.target)
        return
    addr = random.choice(addrs)
    url = "http://%s:%d%s" % (addr[4][0], args.port, args.path)
    s.get(url, timeout=args.timeout)
    logging.info("Did request %s", url)


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

        time.sleep(args.period)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('-target', default="frontend.weave.local")
    parser.add_argument('-port', default=80, type=int)
    parser.add_argument('-path', default="/")
    parser.add_argument('-concurrency', default=1, type=int)
    parser.add_argument('-persist', default=True, type=bool)
    parser.add_argument('-timeout', default=1.0, type=float)
    parser.add_argument('-period', default=0.1, type=float)
    args = parser.parse_args()

    logging.info("Starting %d threads, targeting %s", args.concurrency,
                 args.target)
    threads = [
        threading.Thread(target=do_requests, args=(args, ))
        for i in range(args.concurrency)
    ]
    for thread in threads:
        thread.start()
    for thread in threads:
        thread.join()
    logging.info("Exiting")


if __name__ == "__main__":
    logfmt = '%(asctime)s %(levelname)s %(filename)s:%(lineno)d - %(message)s'
    logging.basicConfig(format=logfmt, level=logging.INFO)
    main()
