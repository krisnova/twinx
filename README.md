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

Use `twinx` to update the title and description on your Twitch channel and your YouTube live broadcast.

```bash 
$ # twinx <integration> update 
$ twinx youtube update
$ twinx twitch update
```

Use `twinx` to start a local `RTMP` router.
This will allow your local OBS to stream to both YouTube and Twitch at the same time.

```bash
$ # Start a local RTMP server
$ twinx rtmp start
```

Use `twinx` to start streaming in OBS via the new RTMP server.

```bash 
$ # Start streaming via OBS
$ twinx obs start
```

Cleanup when you are finished.

```bash 
$ twinx obs stop
$ twinx rtmp stop
$ twinx stream stop
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