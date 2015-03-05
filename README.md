# cort
Distributed services toolkit.

An attempt at base framework for distributed microservices written to improve golang skills and distributed systems knowledge.

Also, made for fun.

## Stuff going to be used (well, probably)

* [Go](http://golang.org/)
* [Consul](http://consul.io/) (Service discovery and coordination)
* [ZeroMQ](http://zeromq.org/) (Brokerless messaging)
* [MessagePack](http://msgpack.org/)

## Introduction

Cort is meant to be used as a framework for independent microservices, providing brokerless communication and service discovery, being easy to use and as reliable as possible.

## Communication

Services exchange messages using ZeroMQ in a brokerless model (peer-to-peer). Some variation of [the Freelance pattern](http://zguide.zeromq.org/page:all#Brokerless-Reliability-Freelance-Pattern) is going to be used.

Some thoughts:

* There's a single ZMQ connection between two services, managed by a goroutine serving as local broker. This goroutine will create two **ROUTER** sockets.
    * First socket, serving as proxy for local clients, binds using **inproc** as transport and the remote service's name as endpoint.
    * Second socket, for outgoing requests and incoming responses, connects to remote service's nodes (discovered before).
    * Local clients use **REQ** sockets.
* On the server side, there will be two **ROUTER** sockets as well.
    * First socket, for incoming requests from remote services and outgoing responses, binds with **TCP** as transport and the address specified by the service as endpoint.
    * Second socket, for worker goroutines, binds with **inproc** as transport and some unique name (known to workers) as endpoint.
    * Workers use **REQ** sockets (for proper load balancing, as proposed in the ZMQ Guide).
* Go's channels could work probably just as well as inproc sockets for local communication. However, it seems like using ZMQ sockets for entire service-to-service communication will make local routing easier and more clear. Although, channels might be more suitable for server workers.

## Discovery

At this time, [Consul](http://consul.io/) is going to be used for services registration and discovery. Various backends should be pluggable in the future, though.

The usage of [Consul Go API](https://github.com/hashicorp/consul/tree/master/api) seems pretty straightforward.

## TODO

Discovery:

* Should connect to a new node after it is discovered. Then send some kind of heartbeat.
* Re-think which module should handle this and how.

## Examples

```go
// to be added
```
