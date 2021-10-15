# Reverse Engineering Twitch RTMP


Original twinx RTMP client (publish)

```
[emma twinx]# tcpdump -nnXSs 0 port 1935 -w twinx.pcap
[emma twinx]# tshark -r twinx.pcap  | grep "RTMP"
Running as user "root" and group "root". This could be dangerous.
    4   0.126650 172.21.137.138 → 136.144.50.161 RTMP 1603 Handshake C0+C1
   10   0.193618 136.144.50.161 → 172.21.137.138 RTMP 243 Handshake S0+S1+S2
   12   0.193889 172.21.137.138 → 136.144.50.161 RTMP 1768 Handshake C2|connect('live')
   13   0.194371 172.21.137.138 → 136.144.50.161 RTMP 176 createStream()|publish('re_1669031_a648c5b23b099e1cb824')
   17   0.261559 136.144.50.161 → 172.21.137.138 RTMP 406 Window Acknowledgement Size 2500000|Set Peer Bandwidth 2500000,Dynamic|Stream Begin 0|Set Chunk Size 512|_result('NetConnection.Connect.Success')
   19   0.264374 136.144.50.161 → 172.21.137.138 RTMP 107 _result()
```

```bash 
[emma twinx]# tcpdump -nnXSs 0 port 1935 -w twinx.pcap
[emma twinx]# tshark -r twinx.pcap  | grep -i rtmp
Running as user "root" and group "root". This could be dangerous.
    7   0.739384 172.21.137.138 → 52.223.243.208 RTMP 1603 Handshake C0+C1
   13   0.812076 52.223.243.208 → 172.21.137.138 RTMP 1602 Handshake S0+S1+S2
   15   0.812471 172.21.137.138 → 52.223.243.208 RTMP 1790 Handshake C2|connect('app')
   16   0.813038 172.21.137.138 → 52.223.243.208 RTMP 190 createStream()|publish('live_733531528_k9ZMBZXSUfOuGCrquQbgeXmLa5Y5ve')
   17   0.825087 52.223.243.208 → 172.21.137.138 RTMP 154 [TCP Spurious Retransmission] |Unknown (0x0)|Unknown (0x0)|Unknown (0x0)|Unknown (0x0)|Unknown (0x0)|Unknown (0x0)
   20   0.884987 52.223.243.208 → 172.21.137.138 RTMP 82 Unknown (0x0)|Unknown (0x0)
   22   0.885048 52.223.243.208 → 172.21.137.138 RTMP 83 Set Peer Bandwidth 2500000,Dynamic
   24   0.885058 52.223.243.208 → 172.21.137.138 RTMP 84 Stream Begin 0
   26   0.885067 52.223.243.208 → 172.21.137.138 RTMP 82 Set Chunk Size 4096
   28   0.885077 52.223.243.208 → 172.21.137.138 RTMP 314 _result('NetConnection.Connect.Success')
   31   0.885093 52.223.243.208 → 172.21.137.138 RTMP 107 _result()
   33   0.885102 52.223.243.208 → 172.21.137.138 RTMP 84 Stream Begin 1
   35   0.940665 52.223.243.208 → 172.21.137.138 RTMP 84 [TCP Spurious Retransmission] |Stream Begin 1

```

NGINX RTMP client (publish)

Note: You must connect with a client such as OBS to get nginx to publish to twitch.

```
    5   0.145993 172.21.130.133 → 52.223.243.22 RTMP 1603 Handshake C0+C1
    9   0.218372 52.223.243.22 → 172.21.130.133 RTMP 1691 Handshake S0+S1+S2
   11   0.218655 172.21.130.133 → 52.223.243.22 RTMP 1602 Handshake C2
   16   0.296317 172.21.130.133 → 52.223.243.22 RTMP 289 Set Chunk Size 4000|Window Acknowledgement Size 5000000|connect('app')
   18   0.364754 52.223.243.22 → 172.21.130.133 RTMP 82 Window Acknowledgement Size 2500000
   20   0.365159 52.223.243.22 → 172.21.130.133 RTMP 83 Set Peer Bandwidth 2500000,Dynamic
   22   0.365203 52.223.243.22 → 172.21.130.133 RTMP 84 Stream Begin 0
   24   0.365214 52.223.243.22 → 172.21.130.133 RTMP 82 Set Chunk Size 4096
   26   0.365223 52.223.243.22 → 172.21.130.133 RTMP 314 _result('NetConnection.Connect.Success')
   28   0.365383 172.21.130.133 → 52.223.243.22 RTMP 103 createStream()
   29   0.415574 52.223.243.22 → 172.21.130.133 RTMP 314 [TCP Spurious Retransmission] |_result('NetConnection.Connect.Success')
   32   0.434101 52.223.243.22 → 172.21.130.133 RTMP 107 _result()
   34   0.434158 52.223.243.22 → 172.21.130.133 RTMP 84 Stream Begin 1
   36   0.434274 172.21.130.133 → 52.223.243.22 RTMP 153 publish('live_733531528_k9ZMBZXSUfOuGCrquQbgeXmLa5Y5ve')
   38   0.883509 172.21.130.133 → 52.223.243.22 RTMP 84 Stream Begin 1
   92   1.874488 52.223.243.22 → 172.21.130.133 RTMP 250 [TCP ZeroWindow] |onStatus('NetStream.Publish.Start')
 1067   4.892199 52.223.243.22 → 172.21.130.133 RTMP 82 Acknowledgement 1250983
 1853   8.637411 52.223.243.22 → 172.21.130.133 RTMP 82 Acknowledgement 2502244
 2653  12.405729 52.223.243.22 → 172.21.130.133 RTMP 82 Acknowledgement 3752456
 3422  16.129172 52.223.243.22 → 172.21.130.133 RTMP 82 Acknowledgement 5003023
 4221  19.891983 52.223.243.22 → 172.21.130.133 RTMP 82 Acknowledgement 6253076
 5027  23.634151 52.223.243.22 → 172.21.130.133 RTMP 82 Acknowledgement 7503346
 5809  27.411475 52.223.243.22 → 172.21.130.133 RTMP 82 Acknowledgement 8753717

```



---

From OBS to nginx server 


```
2021/10/14 12:45:00 [info] 5667#5667: *1 client connected '127.0.0.1'
2021/10/14 12:45:00 [info] 5667#5667: *1 connect: app='twinx' args='' flashver='FMLE/3.0 (compatible; FMSc/1.0)' swf_url='rtmp://localhost:1935/twinx' tc_url='rtmp://localhost:1935/twinx' page_url='' acodecs=0 vcodecs=0 object_encoding=0, client: 127.0.0.1, server: 0.0.0.0:1935
2021/10/14 12:45:00 [info] 5667#5667: *1 createStream, client: 127.0.0.1, server: 0.0.0.0:1935
2021/10/14 12:45:00 [info] 5667#5667: *1 publish: name='live_733531528_k9ZMBZXSUfOuGCrquQbgeXmLa5Y5ve1234' args='' type=live silent=0, client: 127.0.0.1, server: 0.0.0.0:1935
2021/10/14 12:45:08 [info] 5667#5667: *1 deleteStream, client: 127.0.0.1, server: 0.0.0.0:1935
2021/10/14 12:45:08 [info] 5667#5667: *1 disconnect, client: 127.0.0.1, server: 0.0.0.0:1935
2021/10/14 12:45:08 [info] 5667#5667: *1 deleteStream, client: 127.0.0.1, server: 0.0.0.0:1935
```

From Twinx client to nginx server 

```
2021/10/14 12:46:40 [info] 5724#5724: *2 client connected '127.0.0.1'
2021/10/14 12:46:40 [info] 5724#5724: *2 connect: app='twinx' args='' flashver='FMS/3,0,1,123' swf_url='' tc_url='rtmp://localhost:1935/twinx/twinx_XVlBzgbaiCMRAjWwhTHc' page_url='' acodecs=0 vcodecs=0 object_encoding=0, client: 127.0.0.1, server: 0.0.0.0:1935
2021/10/14 12:46:40 [info] 5724#5724: *2 createStream, client: 127.0.0.1, server: 0.0.0.0:1935
2021/10/14 12:46:40 [info] 5724#5724: *2 publish: name='twinx' args='' type=live silent=0, client: 127.0.0.1, server: 0.0.0.0:1935
```

From Improved Twinx client to nginx server 

```
2021/10/14 23:10:51 [info] 2352#2352: *1 client connected '127.0.0.1'
2021/10/14 23:10:51 [info] 2352#2352: *1 connect: app='twinx' args='' flashver='FMS/3,0,1,123' swf_url='rtmp://localhost:1935/twinx' tc_url='rtmp://localhost:1935/twinx' page_url='' acodecs=0 vcodecs=0 object_encoding=0, client: 127.0.0.1, server: 0.0.0.0:1935
2021/10/14 23:10:51 [info] 2352#2352: *1 createStream, client: 127.0.0.1, server: 0.0.0.0:1935
2021/10/14 23:10:51 [info] 2352#2352: *1 publish: name='twinx_XVlBzgbaiCMRAjWwhTHc' args='' type=live silent=0, client: 127.0.0.1, server: 0.0.0.0:1935
2021/10/14 23:10:52 [info] 2352#2352: *1 disconnect, client: 127.0.0.1, server: 0.0.0.0:1935
2021/10/14 23:10:52 [info] 2352#2352: *1 deleteStream, client: 127.0.0.1, server: 0.0.0.0:1935
```