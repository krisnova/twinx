# Goops

Go Operating Proxy Service

An advanced RTMP proxy written in Go and named after cans of sticky things in the garage. 

Designed to support multiple proxy output streams from a single input.

Tee. But for RTMP streams.

```
                                       /-- rtmp --> [TikTok]
                                      /
    [OBS] -- rtmp --> [goops proxy]  +-- rtmp -->   [YouTube]
                                      \
                                       \-- rtmp -->  [Twitch.tv]
```        