Tracer is an prototype for doing container-centric distributed request tracing without applications modifications.

It its very early.  Ask Tom for a demo.

Run tracer:
- make
- ./tracer.sh start

TODO:
- <s>need to stich traces together</s>
- deal with persistent connections
- make it work for goroutines
- test with jvm based app
- find way to get local ip address for an incoming connection
- make the container/process trace start/stop more reliable
