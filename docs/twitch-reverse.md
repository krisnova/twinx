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
  536  12.656952 172.21.137.138 → 52.223.243.160 RTMP 1603 Handshake C0+C1
  539  12.728828 52.223.243.160 → 172.21.137.138 RTMP 3139 Handshake S0+S1+S2
  541  12.728973 172.21.137.138 → 52.223.243.160 RTMP 1602 Handshake C2
  545  12.797736 172.21.137.138 → 52.223.243.160 RTMP 289 Set Chunk Size 4000|Window Acknowledgement Size 5000000|connect('app')
  547  12.866527 52.223.243.160 → 172.21.137.138 RTMP 82 Window Acknowledgement Size 2500000
  549  12.866550 52.223.243.160 → 172.21.137.138 RTMP 83 Set Peer Bandwidth 2500000,Dynamic
  551  12.866554 52.223.243.160 → 172.21.137.138 RTMP 84 Stream Begin 0
  553  12.866556 52.223.243.160 → 172.21.137.138 RTMP 82 Set Chunk Size 4096
  555  12.866765 52.223.243.160 → 172.21.137.138 RTMP 314 _result('NetConnection.Connect.Success')
  557  12.866866 172.21.137.138 → 52.223.243.160 RTMP 103 createStream()
  558  12.899539 52.223.243.160 → 172.21.137.138 RTMP 314 [TCP Spurious Retransmission] |_result('NetConnection.Connect.Success')
  561  12.935569 52.223.243.160 → 172.21.137.138 RTMP 107 _result()
  563  12.935602 52.223.243.160 → 172.21.137.138 RTMP 84 Stream Begin 1
  565  12.935693 172.21.137.138 → 52.223.243.160 RTMP 153 publish('live_733531528_k9ZMBZXSUfOuGCrquQbgeXmLa5Y5ve')
  568  14.529909 52.223.243.160 → 172.21.137.138 RTMP 250 onStatus('NetStream.Publish.Start')

```

