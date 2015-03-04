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

## Discovery

At this time, [Consul](http://consul.io/) is going to be used for services registration and discovery. Various backends should be pluggable in the future, though.

The usage of [Consul Go API](http://github.com/hashicorp/consul/api) seems pretty straightforward.

## Examples

```go
// to be added
```
