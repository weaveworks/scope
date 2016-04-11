FROM ubuntu:wily
MAINTAINER Weaveworks Inc <help@weave.works>
LABEL works.weave.role=system

# Install BCC
RUN apt-key adv --keyserver keyserver.ubuntu.com --recv-keys D4284CDD
RUN echo "deb http://52.8.15.63/apt trusty main" | tee /etc/apt/sources.list.d/iovisor.list
RUN apt-get update && apt-get install -y libbcc libbcc-examples python-bcc

# Add our plugin
ADD ./http-requests.c ./http-requests.py /usr/bin/
ENTRYPOINT ["/usr/bin/http-requests.py"]
