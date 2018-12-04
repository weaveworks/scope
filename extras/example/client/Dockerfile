FROM tatsushid/tinycore-python:2.7
WORKDIR /home/weave
ADD requirements.txt /home/weave/
RUN pip install -r /home/weave/requirements.txt
ADD client.py /home/weave/
ENTRYPOINT ["python", "/home/weave/client.py"]

ARG revision
LABEL maintainer="Weaveworks <help@weave.works>" \
      org.opencontainers.image.title="example-client" \
      org.opencontainers.image.source="https://github.com/weaveworks/scope/tree/master/extras/example/client" \
      org.opencontainers.image.revision="${revision}" \
      org.opencontainers.image.vendor="Weaveworks"
