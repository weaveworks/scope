FROM python:2.7
MAINTAINER Weaveworks Inc <help@weave.works>
ADD . /home/weave
WORKDIR /home/weave
RUN pip install -r /home/weave/requirements.txt
ENTRYPOINT ["python", "/home/weave/app.py"]
