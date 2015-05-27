# example

To run,

```
make
docker-compose up -d
```

# "architecture"

```
curl -> pyapp --> goapp -> elasticsearch
              |
              --> qotd -> internet
              |
              --> redis
```
