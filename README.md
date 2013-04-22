# KLUDGE
## A restful key-value store based on LevelDB

.
|-- common
|-- kludge-backend: the key-value store backend; this is the actual LevelDB
|                   interface.
|-- kludge-client: package simplifying access to a kludge server.
|-- kludge-server: the kludge server front-end that clients communicate with.

## Status

This is still in pre-alpha mode; it's on Github to facilitate code review.
If we're not talking about this, you probably don't want to poke through
it yet.

## The kludge story:

[Ben Johnson](https://github.com/benbjohnson) recently
embarked on a project to port the
[Raft distributed consenus protocol](https://ramcloud.stanford.edu/wiki/download/attachments/11370504/raft.pdf).
Distributed algorithms have long interested me, but I haven't had a
project to work on that requires them (which is the way I learn most
of the things I've learned). Later, I reviewed his slide deck for a
presentation on [SkyDB](http://skydb.io), in which I noticed he had
based his database on [LevelDB](https://code.google.com/p/leveldb/).
I put two and two together, and decided a distributed database would
be interesting to build, and would get me to learn the things I wanted
to learn.
