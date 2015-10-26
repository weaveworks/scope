# SOCKS Proxy

The challenge: you’ve built and deployed your microservices based
application on a Weave network, running on a set of VMs on EC2.  Many
of the services’ public API are reachable from the internet via an
Nginx-based reverse proxy, but some of the services also expose
private monitoring and manage endpoints via embedded HTTP servers.
How do I securely get access to these from my laptop, without exposing
them to the world?

One method we’ve started using at Weaveworks is a 90’s technology - a
SOCKS proxy combined with a PAC script.  It’s relatively
straight-forward: one ssh’s into any of the VMs participating in the
Weave network, starts the SOCKS proxy in a container on Weave the
network, and SSH port forwards a few local port to the proxy.  All
that’s left is for the user to configure his browser to use the proxy,
and voila, you can now access your Docker containers, via the Weave
network (and with all the magic of weavedns), from your laptop’s
browser!

It is perhaps worth noting there is nothing Weave-specific about this
approach - this should work with any SDN or private network.

A quick example:

```
vm1$ weave launch
vm1$ eval $(weave env)
vm1$ docker run -d --name nginx nginx
```

And on your laptop

```
laptop$ git clone https://github.com/weaveworks/tools
laptop$ cd tools/socks
laptop$ ./connect.sh vm1
Starting proxy container...
Please configure your browser for proxy
http://localhost:8080/proxy.pac
```

To configure your Mac to use the proxy:

1. Open System Preferences
2. Select Network
3. Click the 'Advanced' button
4. Select the Proxies tab
5. Click the 'Automatic Proxy Configuration' check box
6. Enter 'http://localhost:8080/proxy.pac' in the URL box
7. Remove `*.local` from the 'Bypass proxy settings for these Hosts & Domains'

Now point your browser at http://nginx.weave.local/
