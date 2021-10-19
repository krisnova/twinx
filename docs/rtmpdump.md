# RTMP Dump (nginx -> VLC)

```
[nova@emma ~]$ rtmpdump -V -v -r "rtmp://localhost:1935/twinx/1234" -o - | "vlc" -
VLC media player 3.0.16 Vetinari (revision 3.0.13-8-g41878ff4f2)
RTMPDump v2.4
(c) 2010 Andrej Stepanchuk, Howard Chu, The Flvstreamer Team; license: GPL
DEBUG: Parsing...
DEBUG: Parsed protocol: 0
DEBUG: Parsed host    : localhost
DEBUG: Parsed app     : twinx
DEBUG: Protocol : RTMP
DEBUG: Hostname : localhost
DEBUG: Port     : 1935
DEBUG: Playpath : 1234
DEBUG: tcUrl    : rtmp://localhost:1935/twinx
DEBUG: app      : twinx
DEBUG: live     : yes
DEBUG: timeout  : 30 sec
DEBUG: Setting buffer time to: 36000000ms
Connecting ...

DEBUG: RTMP_Connect1, ... connected, handshaking
DEBUG: HandShake: Type Answer   : 03
DEBUG: HandShake: Server Uptime : 1433139062
DEBUG: HandShake: FMS Version   : 0.0.0.0
DEBUG: HandShake: Handshaking finished....
DEBUG: RTMP_Connect1, handshaked
DEBUG: Invoking connect
INFO: Connected...

DEBUG: HandleServerBW: server BW = 5000000
DEBUG: HandleClientBW: client BW = 5000000 2
DEBUG: HandleChangeChunkSize, received: chunk size change to 4000
DEBUG: RTMP_ClientPacket, received: invoke 190 bytes
DEBUG: (object begin)
DEBUG: (object begin)
DEBUG: Property: <Name:             fmsVer, STRING:     FMS/3,0,1,123>
DEBUG: Property: <Name:       capabilities, NUMBER:     31.00>
DEBUG: (object end)
DEBUG: (object begin)
DEBUG: Property: <Name:              level, STRING:     status>
DEBUG: Property: <Name:               code, STRING:     NetConnection.Connect.Success>
DEBUG: Property: <Name:        description, STRING:     Connection succeeded.>
DEBUG: Property: <Name:     objectEncoding, NUMBER:     0.00>
DEBUG: (object end)
DEBUG: (object end)
DEBUG: HandleInvoke, server invoking <_result>
DEBUG: HandleInvoke, received result for method call <connect>
DEBUG: sending ctrl. type: 0x0003

DEBUG: Invoking createStream
DEBUG: FCSubscribe: 1234
DEBUG: Invoking FCSubscribe
DEBUG: RTMP_ClientPacket, received: invoke 29 bytes
DEBUG: (object begin)
DEBUG: Property: NULL
DEBUG: (object end)
DEBUG: HandleInvoke, server invoking <_result>
DEBUG: HandleInvoke, received result for method call <createStream>
DEBUG: SendPlay, seekTime=0, stopTime=0, sending play: 1234
DEBUG: Invoking play
DEBUG: sending ctrl. type: 0x0003
DEBUG: RTMP_ClientPacket, received: invoke 96 bytes
DEBUG: (object begin)
DEBUG: Property: NULL
DEBUG: (object begin)
DEBUG: Property: <Name:              level, STRING:     status>
DEBUG: Property: <Name:               code, STRING:     NetStream.Play.Start>
DEBUG: Property: <Name:        description, STRING:     Start live>
DEBUG: (object end)
DEBUG: (object end)
DEBUG: HandleInvoke, server invoking <onStatus>
DEBUG: HandleInvoke, onStatus: NetStream.Play.Start
Starting Live Stream
DEBUG: RTMP_ClientPacket, received: notify 24 bytes
DEBUG: (object begin)
DEBUG: (object end)
^CCaught signal: 2, cleaning up, just a second...
DEBUG: RTMPSockBuf_Fill, recv returned -1. GetSockError(): 4 (Interrupted system call)
DEBUG: Invoking deleteStream
ERROR: RTMP_ReadPacket, failed to read RTMP packet header
-0.001 kB / 0.00 sec
DEBUG: RTMP_Read returned: 0
Download may be incomplete (downloaded about 0.00%), try resuming
DEBUG: Closing connection.

QObject::~QObject: Timers cannot be stopped from another thread
```

# RTMP Dump (Twinx -> VLC)

```
[nova@emma ~]$ rtmpdump -V -v -r "rtmp://localhost:1935/twinx/1234" -o - | "vlc" -
VLC media player 3.0.16 Vetinari (revision 3.0.13-8-g41878ff4f2)
RTMPDump v2.4
(c) 2010 Andrej Stepanchuk, Howard Chu, The Flvstreamer Team; license: GPL
DEBUG: Parsing...
DEBUG: Parsed protocol: 0
DEBUG: Parsed host    : localhost
DEBUG: Parsed app     : twinx
DEBUG: Protocol : RTMP
DEBUG: Hostname : localhost
DEBUG: Port     : 1935
DEBUG: Playpath : 1234
DEBUG: tcUrl    : rtmp://localhost:1935/twinx
DEBUG: app      : twinx
DEBUG: live     : yes
DEBUG: timeout  : 30 sec
DEBUG: Setting buffer time to: 36000000ms
Connecting ...

DEBUG: RTMP_Connect1, ... connected, handshaking
DEBUG: HandShake: Type Answer   : 03
DEBUG: HandShake: Server Uptime : 0
DEBUG: HandShake: FMS Version   : 0.0.0.0
DEBUG: HandShake: Handshaking finished....
DEBUG: RTMP_Connect1, handshaked
DEBUG: Invoking connect
INFO: Connected...

DEBUG: RTMP_ClientPacket, received: bytes read report
DEBUG: HandleClientBW: client BW = 2500000 2
DEBUG: HandleChangeChunkSize, received: chunk size change to 8192
DEBUG: RTMP_ClientPacket, received: invoke 191 bytes
DEBUG: (object begin)
DEBUG: (object begin)
DEBUG: Property: <Name:             fmsVer, STRING:     LNX 10,0,32,18>
DEBUG: Property: <Name:       capabilities, NUMBER:     31.00>
DEBUG: (object end)
DEBUG: (object begin)
DEBUG: Property: <Name:              level, STRING:     status>
DEBUG: Property: <Name:               code, STRING:     NetConnection.Connect.Success>
DEBUG: Property: <Name:        description, STRING:     Connection succeeded.>
DEBUG: Property: <Name:     objectEncoding, NUMBER:     0.00>
DEBUG: (object end)
DEBUG: (object end)
DEBUG: HandleInvoke, server invoking <_result>
DEBUG: HandleInvoke, received result for method call <connect>
DEBUG: sending ctrl. type: 0x0003

DEBUG: Invoking createStream
DEBUG: FCSubscribe: 1234
DEBUG: Invoking FCSubscribe
DEBUG: RTMP_ClientPacket, received: invoke 29 bytes
DEBUG: (object begin)
DEBUG: Property: NULL
DEBUG: (object end)
DEBUG: HandleInvoke, server invoking <_result>
DEBUG: HandleInvoke, received result for method call <createStream>
DEBUG: SendPlay, seekTime=0, stopTime=0, sending play: 1234
DEBUG: Invoking play
DEBUG: sending ctrl. type: 0x0003
DEBUG: HandleCtrl, received ctrl. type: 4, len: 6
DEBUG: HandleCtrl, Stream IsRecorded 1
DEBUG: HandleCtrl, received ctrl. type: 0, len: 6
DEBUG: HandleCtrl, Stream Begin 1
DEBUG: RTMP_ClientPacket, received: invoke 115 bytes
DEBUG: (object begin)
DEBUG: Property: NULL
DEBUG: (object begin)
DEBUG: Property: <Name:               code, STRING:     NetStream.Play.Reset>
DEBUG: Property: <Name:        description, STRING:     Playing and resetting stream.>
DEBUG: Property: <Name:              level, STRING:     status>
DEBUG: (object end)
DEBUG: (object end)
DEBUG: HandleInvoke, server invoking <onStatus>
DEBUG: HandleInvoke, onStatus: NetStream.Play.Reset
DEBUG: RTMP_ClientPacket, received: invoke 109 bytes
DEBUG: (object begin)
DEBUG: Property: NULL
DEBUG: (object begin)
DEBUG: Property: <Name:        description, STRING:     Started playing stream.>
DEBUG: Property: <Name:              level, STRING:     status>
DEBUG: Property: <Name:               code, STRING:     NetStream.Play.Start>
DEBUG: (object end)
DEBUG: (object end)
DEBUG: HandleInvoke, server invoking <onStatus>
DEBUG: HandleInvoke, onStatus: NetStream.Play.Start
Starting Live Stream
DEBUG: RTMP_ClientPacket, received: invoke 109 bytes
DEBUG: (object begin)
DEBUG: Property: NULL
DEBUG: (object begin)
DEBUG: Property: <Name:        description, STRING:     Started playing stream.>
DEBUG: Property: <Name:              level, STRING:     status>
DEBUG: Property: <Name:               code, STRING:     NetStream.Data.Start>
DEBUG: (object end)
DEBUG: (object end)
DEBUG: HandleInvoke, server invoking <onStatus>
DEBUG: HandleInvoke, onStatus: NetStream.Data.Start
DEBUG: RTMP_ClientPacket, received: invoke 113 bytes
DEBUG: (object begin)
DEBUG: Property: NULL
DEBUG: (object begin)
DEBUG: Property: <Name:              level, STRING:     status>
DEBUG: Property: <Name:               code, STRING:     NetStream.Publish.Notify>
DEBUG: Property: <Name:        description, STRING:     Started playing notify.>
DEBUG: (object end)
DEBUG: (object end)
DEBUG: HandleInvoke, server invoking <onStatus>
DEBUG: HandleInvoke, onStatus: NetStream.Publish.Notify
^CCaught signal: 2, cleaning up, just a second...
DEBUG: RTMPSockBuf_Fill, recv returned -1. GetSockError(): 4 (Interrupted system call)
DEBUG: Invoking deleteStream
ERROR: RTMP_ReadPacket, failed to read RTMP packet header
-0.001 kB / 0.00 sec
DEBUG: RTMP_Read returned: 0
Download may be incomplete (downloaded about 0.00%), try resuming
DEBUG: Closing connection.

QObject::~QObject: Timers cannot be stopped from another thread

```