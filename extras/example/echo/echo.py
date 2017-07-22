import logging

from flask import Flask
from flask import request
from werkzeug.serving import WSGIRequestHandler

app = Flask(__name__)


@app.route('/')
def echo():
    return request.data


if __name__ == "__main__":
    logfmt = '%(asctime)s %(levelname)s %(filename)s:%(lineno)d - %(message)s'
    logging.basicConfig(format=logfmt, level=logging.INFO)
    WSGIRequestHandler.protocol_version = "HTTP/1.0"
    app.run(host="0.0.0.0", port=80, debug=True)
