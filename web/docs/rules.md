# Rules

Rules are used to classify (group and color) hosts, and to filter edges
either in or out.

The idea is you make a rule to match hosts by their traffic, or you make a
rule to match addresses which talk to a host.

## Basic syntax

A basic rule has this structure, where everything is optional:

```
Rule     := <not> <director> <selector>
Director := ingress | egress | remote-ingress | remote-egress
Selector := tcp <port> | udp <port> | ip | ipv6 | icmp | icmpv6 | arp
```

Rules are case insensitive.

### Directors

Directors tell what the selector applies on. If not specified it'll match any.

- `ingress`: measured ingress traffic. Must be a probe.
- `egress`: measured egress traffic. Must be a probe.
- `remote-ingress`: ingress on the remote side of an edge. The local side may
  or may not be a probe. Use this to select addresses which talk to a probe,
  but don't need to be one.
- `remote-egress`: same idea as 'remote-ingress', but egress traffic.

`not` can be used to inverse the match.

### Selectors

selectors tell which kind of traffic matches. For `tcp` and `udp` it is
possible to specify ports.

- `tcp <<src|dst> port>`: TCP traffic. If a port is given it should either come
  from or go to that port. If `src` or `dst` is given that'll apply to the port
  number.
- `ucp <<src|dst> port>`: Same idea as `tcp`, but for UDP traffic.
- `ip`, `ipv4`: any IPv4 traffic
- `ipv6`: any IPv6 traffic
- `icmp` or `icmpv4`: any ICMP traffic
- `icmpv6`: any ICMPv6 traffic.

## Precedence

You can compose rules with 'and' and 'or' ('and' has higher precedence), or use
parentheses.

```
<rule> [and|or] <rule>
```

Examples.

- `tcp and udp`
- `tcp or udp`
- `tcp 80 or tcp 8080`
- `tcp 80 or tcp 8080`
- `egress tcp dst 80 and egress tcp src 80`
- `(egress tcp dst 80 or egress tcp dst 8080) and egress tcp src 80`

# Sample rules

Rule                                                     | Description
---------------------------------------------------------|------------
`tcp`                                                    | Any TCP traffic on the edge
`tcp 80`                                                 | Any TCP traffic related to port 80
`tcp and ipv4`                                           | Edge has TCP over IPv4
`ip or ipv6`                                             | Any IP traffic
`src 80`                                                 | TCP or UDP traffic coming from port 80, either on the sending or the receiving side
`egress src 80`                                          | TCP or UDP traffic being send from port 80
`tcp or udp`                                             | Any TCP or traffic on the edge
`ingress tcp dst 80`                                     | Traffic going in to port 80. Likely a web server, but this rule also matches traffic to ports where nothing listens on
`ingress tcp dst 80` `and egress src 80`                   | Traffic going in to port 80, and traffic going out from it. This is a functioning TCP server on port 80
`ingress TCP dst 80` `and not egress TCP src 80`           | Traffic going in, but there is no reply. This is a broken web server
`egress tcp dst 80` ` and ingress src 80`                   | A probe which is a HTTP client
`remote-ingress tcp dst 80` `and remote-egress tcp src 80` | Something which talks to a probe with a port 80 server. It doesn't matter if there is a probe installed or not

