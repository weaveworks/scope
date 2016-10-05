# Start from weaveworks/scope, so that we have a docker client built in.
FROM python:2
MAINTAINER Weaveworks Inc <help@weave.works>
LABEL works.weave.role=system

RUN pip install docker-py

# Add our plugin
ADD ./volume-count.py /usr/bin/volume-count.py
ENTRYPOINT ["/usr/bin/volume-count.py"]
