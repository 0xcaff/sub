push-sub
========

A daemon which maintains subscriptions with PubSubHubbub (PuSH) hubs and calls
specified commands when messages arrive. It aims to make receiving PuSH updates
secure and easy.

[![thumb]][video]

Usage
-----

Create a config file and run `push-sub`. Here's an example config file:

```toml
# config.toml
address="127.0.0.1:17889"
basepath="https://hub.ydns.eu/"

[subscriptions.youtube_channel]
topic="https://www.youtube.com/xml/feeds/videos.xml?channel_id=UCF8wOZvgrBZzj4netRo3m2A"
hub="https://pubsubhubbub.appspot.com"
command=["/usr/bin/tee", "/tmp/pub.txt"]
```

For more configuration options, check out [`config.go`][config].

Now run `push-sub` and wait for messages. Make sure that `basepath` is
publically visible and points to `address`. When a message arrives, `command`
will be executed and the message will passed to it through standard input.

    INFO[0000] reading config: ./config.toml                
    INFO[0000] discovering hub                                  name="youtube_channel"
    INFO[0000] discovered hub: https://pubsubhubbub.appspot.com name="youtube_channel"
    INFO[0000] registered endpoint="https://hub.ydns.eu/ZC8IRIuuaObmkaOx1zC8NhxjGDJ0E6n4FT7TjMAyyot3f6zxvupGacl56isMR6rGkHebwRwitNXObtAzjYxBKt0ze3k9XTD9q1i" name="youtube_channel"
    INFO[0000] listening on 127.0.0.1:17889
    INFO[0001] subscribing                                   name="youtube_channel"
    INFO[0001] subscribed                                    name="youtube_channel"
    INFO[0013] received message                              name="youtube_channel"
    INFO[0013] running command                               args=[/usr/bin/tee /tmp/pub.txt] name="youtube_channel"

Installing
----------

    go get github.com/0xcaff/sub/push-sub

If you don't have `go`, check out the releases for prebuilt binaries.

Daemonizing (Systemd)
-----------

    [Unit]
    Description=PubSubHubbub Subscriber
    Requires=network-online.target

    [Service]
    Type=simple
    ExecStart=/opt/push-sub -config /opt/push-sub.toml

    [Install]
    WantedBy=multi-user.target

[releases]: https://github.com/0xcaff/sub/releases
[thumb]: https://i.imgur.com/tMr6WMv.png
[video]: https://asciinema.org/a/124256
[config]: https://github.com/0xcaff/sub/blob/master/push-sub/config.go
