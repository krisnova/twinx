# RTMP Dump (nginx -> VLC)

```
DEBUG: HandleCtrl, received ctrl. type: 0, len: 6
DEBUG: HandleCtrl, Stream Begin 1
DEBUG: RTMP_ClientPacket, received: notify 387 bytes
DEBUG: (object begin)
DEBUG: (object begin)
DEBUG: Property: <Name:             Server, STRING:     NGINX RTMP (github.com/arut/nginx-rtmp-module)>
DEBUG: Property: <Name:              width, NUMBER:     1280.00>
DEBUG: Property: <Name:             height, NUMBER:     720.00>
DEBUG: Property: <Name:       displayWidth, NUMBER:     1280.00>
DEBUG: Property: <Name:      displayHeight, NUMBER:     720.00>
DEBUG: Property: <Name:           duration, NUMBER:     0.00>
DEBUG: Property: <Name:          framerate, NUMBER:     30.00>
DEBUG: Property: <Name:                fps, NUMBER:     30.00>
DEBUG: Property: <Name:      videodatarate, NUMBER:     2500.00>
DEBUG: Property: <Name:       videocodecid, NUMBER:     7.00>
DEBUG: Property: <Name:      audiodatarate, NUMBER:     160.00>
DEBUG: Property: <Name:       audiocodecid, NUMBER:     10.00>
DEBUG: Property: <Name:            profile, STRING:     >
DEBUG: Property: <Name:              level, STRING:     >
DEBUG: (object end)
DEBUG: (object end)
INFO: Metadata:
INFO:   Server                NGINX RTMP (github.com/arut/nginx-rtmp-module)
INFO:   width                 1280.00
INFO:   height                720.00
INFO:   displayWidth          1280.00
INFO:   displayHeight         720.00
INFO:   duration              0.00
INFO:   framerate             30.00
INFO:   fps                   30.00
INFO:   videodatarate         2500.00
INFO:   videocodecid          7.00
INFO:   audiodatarate         160.00
INFO:   audiocodecid          10.00
481.393 kB / 1.53 sec[00007f63280063d0] gl gl: Initialized libplacebo v3.120.3 (API v120)
[00007f63280063d0] glconv_vaapi_x11 gl error: vaInitialize: unknown libva error
libva error: vaGetDriverNameByIndex() failed with unknown libva error, driver_name = (null)
[00007f63280063d0] glconv_vaapi_drm gl error: vaInitialize: unknown libva error
libva error: vaGetDriverNameByIndex() failed with unknown libva error, driver_name = (null)
[00007f63280063d0] glconv_vaapi_drm gl error: vaInitialize: unknown libva error
546.725 kB / 1.73 sec[00007f63280063d0] gl gl: Initialized libplacebo v3.120.3 (API v120)
[00007f635006f510] avcodec decoder: Using NVIDIA VDPAU Driver Shared Library  470.74  Mon Sep 13 22:58:37 UTC 2021 for hardware decoding
612.199 kB / 1.94 sec[live_flv @ 0x7f6350076480] Packet mismatch -1651214103 10434 459148
1669.267 kB / 5.17 sec                                                                      
DEBUG: HandleCtrl, received ctrl. type: 1, len: 6
DEBUG: HandleCtrl, Stream EOF 1
^CCaught signal: 2, cleaning up, just a second...
DEBUG: RTMPSockBuf_Fill, recv returned -1. GetSockError(): 4 (Interrupted system call)
DEBUG: Invoking deleteStream
[00007f635c000c80] main input error: ES_OUT_SET_(GROUP_)PCR  is called too late (pts_delay increased to 1000 ms)ERROR: RTMP_ReadPacket, failed to read RTMP packet header               
                                                                                            
1669.689 kB / 5.18 sec
DEBUG: RTMP_Read returned: 0
Download may be incomplete (downloaded about 0.00%), try resuming
DEBUG: Closing connection.

[00007f63507dc2b0] main decoder error: Timestamp conversion failed for 5162001: no reference clock                                                                                      
[00007f63507dc2b0] main decoder error: Could not convert timestamp 0 for faad
QObject::~QObject: Timers cannot be stopped from another thread
```

# RTMP Dump (Twinx -> VLC)

```

```