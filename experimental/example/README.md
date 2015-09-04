# example

To run,

```
make all
./run.sh
```

# "architecture"

```
curl -> frontend --> app --> searchapp -> elasticsearch
         (nginx) |
                 --> qotd -> internet
                 |
                 --> redis
```
