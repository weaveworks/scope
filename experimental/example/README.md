# example

To run,

```
make all
./run.sh
```

# "architecture"

```
curl -> app --> searchapp -> elasticsearch
            |
            --> qotd -> internet
            |
            --> redis
```
