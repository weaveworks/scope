# example

To run,

```
make run
```

# "architecture"

```
curl -> pyapp (x2) --> goapp (x2) -> elasticsearch (x3)
                  |
                   --> qotd -> internet
                  |
                   --> redis
```
