## Overview

The dialer scripts can be used to test Scope with a high number of
containers and connections.

The dialer server is a TCP server that holds incoming connections
forever.

The dialer client is a TCP client that opens n connection to a server
and holds them forever.

The `listener` script starts a dialer server and prints its address for
usage with the `dialer` script. The `dialer` scripts starts up to n
(default 50) client containers, each opening a random (1-20) number of
connections.

`time-scope-probe` then can be used to measure the scheduled time
(utime + stime) of the scope-probe process on the host. The results
can be used to compare performance under different scenarios/branches.

## Usage example

```
# Start a listener
./tools/dialer/listener
Listening on :8082
IP addr + port: 172.17.0.2:8082


# Start the dialer script with a maximum of 10 dialer containers
# (default 50)
./tools/dialer/dialer 172.17.0.2:8082 10


# Start time-scope-probe to measure the scheduled time of scope-probe
# every 3 seconds (default 10 seconds) for 3 times (default 60 times)
sudo ./tools/dialer/time-scope-probe 3 3
...
```

## Build dialer container

```
go build -o bin/dialer
docker build -t dialer .
```
