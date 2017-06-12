sub
===

[![Build Status][build-status-image]][build-status]

Sub is a go library to consume PubSubHubbub (PuSH) hubs. PuSH is a
server-to-server update protocol. It's used to notify servers about changes on
other servers. Check out the [documentation] for examples of using this library.

If you don't need fine-grained control of your subscriptions, check out
[push-sub]. It's a daemon which maintains subscriptions and calls specified
commands when messages arrive.

[documentation]: https://godoc.org/github.com/0xcaff/sub
[push-sub]: https://github.com/0xcaff/sub/tree/master/push-sub
[build-status]: https://travis-ci.org/0xcaff/sub
[build-status-image]: https://travis-ci.org/0xcaff/sub.svg?branch=master
