FROM tatsushid/tinycore-python:2.7
LABEL maintainer="Weaveworks Inc <help@weave.works>"
WORKDIR /home/weave
ADD requirements.txt /home/weave/
RUN pip install -r /home/weave/requirements.txt
ADD echo.py /home/weave/
EXPOSE 5000
ENTRYPOINT ["python", "/home/weave/echo.py"]
