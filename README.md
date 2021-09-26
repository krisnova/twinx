# twinx

```
  ████████╗██╗    ██╗██╗███╗   ██╗██╗  ██╗
  ╚══██╔══╝██║    ██║██║████╗  ██║╚██╗██╔╝
     ██║   ██║ █╗ ██║██║██╔██╗ ██║ ╚███╔╝
     ██║   ██║███╗██║██║██║╚██╗██║ ██╔██╗
     ██║   ╚███╔███╔╝██║██║ ╚████║██╔╝ ██╗
     ╚═╝    ╚══╝╚══╝ ╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝
```

`twinx` is a live-streaming command line tool for Linux. 
It connects streaming services (like Twitch, OBS and YouTube) together via a common `title` and `description`.

### Overview

Use `twinx` to manage your streams.
This should be the first and last tool you use while live-streaming.
This will kick off a persistent standalone linux process that runs as a temporary daemon for the duration of the stream.
The daemon can then be interfaced with using the `twinx` command line tool.
You can add integrations and functionality to your stream after it has been started.

## Example

Use `twinx` to start a new stream on a Linux filesystem.
This will register your `title` and `description` for your new stream, and start the background process.

```bash 
$ # twinx stream start <title> <description>
$ twinx stream start \
    "How to hack the planet" \
    "This is a live stream about how to hack the kernel to burn down capitalism"
```

Start an RTMP Relay. If no `host:port` is defined, `twinx` will select a port and listen on `localhost`.

```bash 
$ twinx relay start <optional host:port>
$ twinx relay start localhost:1719
```



## Configuration

Twitch Callback URL Port: 1717

Environmental Variables

```bash
# OBS
export TWINX_OBS_PORT="1718"
export TWINX_OBS_PASSWORD=""
export TWINX_OBS_HOST="localhost"

# YouTube
export TWINX_YOUTUBE_API_KEY=""
```

## Permissions

`twinx` will manage RTMP servers, unix sockets, and gRPC servers and clients for you.
`twinx` requires root privileges to do this.

## FAQ

This tool is a reflection of being asked the following questions

> What are you working on today?
> Where is that video you did on <thing>
> What is this stream even about?


## Installing

Build the binary.

Install the binary.

```
./configure
make
sudo make install
```

Arch Linux dependencies

```
protoc-gen-go protoc-gen-go-grpc
```

#### RTMP Protocol Reference

 - [Wikipedia Page](https://en.wikipedia.org/wiki/Real-Time_Messaging_Protocol)
 - [How RTMP Works](https://ottverse.com/rtmp-real-time-messaging-protocol-encoding-streaming/)
 - [What is RTMP](https://blog.pogrebnyak.info/what-is-rtmp-and-how-its-used-in-live-streaming/)
 - [RTMP Spec](https://wwwimages2.adobe.com/content/dam/acom/en/devnet/rtmp/pdf/rtmp_specification_1.0.pdf)
 - [NGINX RTMP Module Handshake C Source](https://github.com/arut/nginx-rtmp-module/blob/master/ngx_rtmp_handshake.c)
 - [LiveGo](https://github.com/gwuhaolin/livego/tree/master/protocol/rtmp)