.PHONY: docker
docker: runsvinit
	docker build -t runsvinit-example .

runsvinit: runsvinit-linux-amd64.tgz
	tar zxf $<

runsvinit-linux-amd64.tgz:
	wget --quiet https://github.com/peterbourgon/runsvinit/releases/download/v2.0.0/runsvinit-linux-amd64.tgz

