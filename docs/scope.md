
## Multi host setup

Weave Scope uses WeaveDNS to automatically discover other instances of Scope running on your network.  If you have a running WeaveDNS setup, you do not need any further steps.

If you do not wish to use WeaveDNS, you can instruct Scope to cluster with other Scope instances on the command line.  Hostnames and IP addresses are acceptable, both with and without ports:

```
# weave launch scope1:4030 192.168.0.12 192.168.0.11:4030
```
