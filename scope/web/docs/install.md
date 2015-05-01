# Getting Cello

## Download

We provide Cello as a single archive, containing all daemons for every
supported platform.

All binaries are statically linked and completely self-contained. The only
exception is the probe, which depends on libpcap.

## Demo

After downloading and unarchiving, you use the `./run.sh` script to start a
local Cello stack and play around.

This will give you the traffic of the machine you run Cello on. To complete the
network graph, download and unarchive Cello on every server you want to include
and run `./run.sh --probe-only` there (make sure incoming connections to 4030
are allowed). Cello will find those new servers by itself, no further
configuration needed.

# Installing Cello

## Daemons

Cello has three daemons.

- **cello-app** is the main interface. You just need one.

- **cello-bridge** acts as a bridge between the main interface and probes.
  In most networks, you'll just need one.

- **cello-probe** sniffs traffic and produces reports. Ideally, you'll have
  a probe installed on every machine in your network.

## General procedure

To install a binary on a host, copy it to a canonical path, and use your
process supervision system to start and manage it.

If you don't already have a canonical path for applications like Cello, we
suggest `/usr/local/cello`. If you don't already have a process supervision
system, we suggest [runit](http://smarden.org/runit).

If you use Debian or Ubuntu, and would prefer to use standard .deb packages to
install Cello, we provide a .deb for each binary in the `debian` subdirectory
of the archive.

## Installation

1. Install **cello-app**, **cello-bridge**, and **cello-probe** on the same
host. Let's say that host has the hostname `monitor`.

2. Install **cello-probe** on all other hosts in your network you'd like to
monitor as well.  The cello-bridge will find those cello-probes automatically so
you don't need to configure them.  Make sure incoming connections to 4030 are
allowed by any firewall.

3. Point your web browser to http://monitor:4040. That's it!

### Probe discovery

By default, the `cello-bridge` daemon will parse all incoming probe reports for
remote addresses. For any address that falls in the bridge's network(s), the
bridge will continuously make probe connection attempts on port 4030. In this
way, the whole cello system is configuration-free.

The probe discovery can be disabled with the `-discovery-enabled=false` flag to
`cello-bridge`.  If you disable the probe discovery you will need to supply all
probe addresses with the `-probe-addresses=1.2.3.4:4030,4.5.6.7:4030` argument
to `cello-bridge`.

### Application level statistics

Cello can provide statistics for supported application level protocols.
Enabling this is port based, and is done via the `-dpi=` flag to `cello-probe`.
By default all supported protocols are enabled on their default port, but if
you use non-default ports you need to use the `-dpi=` flag. See `cello-probe
--help` for the up-to-date list of supported protocols, and which ports they
are associated with.

### Scaling to larger networks

For large installations, it may make sense to break your network into logical
segments. You can install one bridge per segment, and configure all probes in
that segment to talk to that bridge. Then, you can configure cello-app to
talk to multiple bridges, via `-bridges=host1:4030,host2:4030,host3:4030`.
