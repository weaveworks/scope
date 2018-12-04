FROM tatsushid/tinycore-python:2.7
WORKDIR /home/weave
ADD requirements.txt /home/weave/
RUN pip install -r /home/weave/requirements.txt
ADD app.py /home/weave/
EXPOSE 5000
ENTRYPOINT ["python", "/home/weave/app.py"]

ARG revision
LABEL maintainer="Weaveworks <help@weave.works>" \
      org.opencontainers.image.title="example-app" \
      org.opencontainers.image.source="https://github.com/weaveworks/scope/tree/master/extras/example/app" \
      org.opencontainers.image.revision="${revision}" \
      org.opencontainers.image.vendor="Weaveworks"
