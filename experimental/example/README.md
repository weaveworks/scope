# example

To run a on a Mac, run:

```
$ docker-machine create -d virtualbox --virtualbox-memory=4096 scope-tastic
$ eval $(docker-machine env scope-tastic)
$ sudo curl -L git.io/weave -o /usr/local/bin/weave
$ sudo chmod +x /usr/local/bin/weave
$ weave launch
$ curl -o run.sh https://raw.githubusercontent.com/weaveworks/scope/master/experimental/example/run.sh
$ ./run.sh
$ sudo wget -O /usr/local/bin/scope https://github.com/weaveworks/scope/releases/download/latest_release/scope
$ sudo chmod a+x /usr/local/bin/scope
$ scope launch
```

# "architecture"

```
curl -> frontend --> app --> searchapp -> elasticsearch
         (nginx) |
                 --> qotd -> internet
                 |
                 --> redis
```
