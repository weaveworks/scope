FROM tatsushid/tinycore-python:2.7
LABEL maintainer="Weaveworks Inc <help@weave.works>"
WORKDIR /home/weave
ADD requirements.txt /home/weave/
RUN pip install -r /home/weave/requirements.txt
ADD client.py /home/weave/
ENTRYPOINT ["python", "/home/weave/client.py"]
