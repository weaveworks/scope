[![Build Status](https://travis-ci.org/paypal/ionet.png)](https://travis-ci.org/paypal/ionet)

ionet provides a [net.Conn](http://golang.org/pkg/net/#Conn) and a
[net.Listener](http://golang.org/pkg/net/#Listener) in which connections
use an [io.Reader](http://golang.org/pkg/io/#Reader) and an
[io.Writer](http://golang.org/pkg/io/#Writer) instead of a traditional
network stack.

This can be handy in unit tests, because it enables you to mock out
the network.

It's also useful when using an external network stack. At PayPal, ionet
is used in [PayPal Beacon](https://www.paypal.com/beacon). Beacon
uses a Bluetooth Low Energy chip accessed over a serial connection.
ionet enables the use of net-based code, such as the
stdlib's [net/http]((http://golang.org/pkg/net/http/), with a
mediated network.

`go get github.com/paypal/ionet`

See godoc for usage.

ionet requires Go 1.1 or later, and is released under a BSD-style license similar to Go's.
