# Publish Response Bug

October, 2021

---

# (Healthy) Nginx PCAP

Using the NGINX module for [RTMP](https://github.com/arut/nginx-rtmp-module) we were able to `tcpdump` a working/healthy OBS transaction.

```bash
[root@emily]: ~/twinx># tshark -r nginx.pcap | grep -i rtmp | grep -v Data
Running as user "root" and group "root". This could be dangerous.
    4   0.000070    127.0.0.1 → 127.0.0.1    RTMP 1603 Handshake C0+C1
    8   0.000119    127.0.0.1 → 127.0.0.1    RTMP 1602 Handshake C2
    9   0.000120    127.0.0.1 → 127.0.0.1    RTMP 1602 Handshake S0+S1+S2
   12   0.000134    127.0.0.1 → 127.0.0.1    RTMP 82 Set Chunk Size 4096
   14   0.000143    127.0.0.1 → 127.0.0.1    RTMP 252 connect('twinx')
   16   0.000171    127.0.0.1 → 127.0.0.1    RTMP 82 Wind ow Acknowledgement Size 5000000
   18   0.000181    127.0.0.1 → 127.0.0.1    RTMP 83 Set Peer Bandwidth 5000000,Dynamic
   20   0.000188    127.0.0.1 → 127.0.0.1    RTMP 82 Set Chunk Size 4000
   22   0.000198    127.0.0.1 → 127.0.0.1    RTMP 268 _result('NetConnection.Connect.Success')
   24   0.000217    127.0.0.1 → 127.0.0.1    RTMP 107 releaseStream('1234')
   26   0.000227    127.0.0.1 → 127.0.0.1    RTMP 103 FCPublish('1234')
   28   0.000234    127.0.0.1 → 127.0.0.1    RTMP 99 createStream()
   30   0.000251    127.0.0.1 → 127.0.0.1    RTMP 107 _result()
   32   0.000265    127.0.0.1 → 127.0.0.1    RTMP 112 publish('1234')
   34   0.000303    127.0.0.1 → 127.0.0.1    RTMP 183 onStatus('NetStream.Publish.Start')
   46   0.783002    127.0.0.1 → 127.0.0.1    RTMP 4163 Unknown (0x0)
   52   0.783035    127.0.0.1 → 127.0.0.1    RTMP 4163 Unknown (0x0)
   54   0.783042    127.0.0.1 → 127.0.0.1    RTMP 4163 Unknown (0x0)
  940   4.982009    127.0.0.1 → 127.0.0.1    RTMP 105 FCUnpublish()
  941   4.982030    127.0.0.1 → 127.0.0.1    RTMP 108 deleteStream()
  943   4.982571    127.0.0.1 → 127.0.0.1    RTMP 186 onStatus('NetStream.Unpublish.Success')
```

# Twinx PCAP

### (Unhealthy)

Before [65926c904a1bd66a9d07d49e8d820ac012ab4676](https://github.com/kris-nova/twinx/commit/65926c904a1bd66a9d07d49e8d820ac012ab4676)

```bash
tshark -r twinx.pcap  | grep -i RTMP | grep -v Data
    4   0.000131    127.0.0.1 → 127.0.0.1    RTMP 1603 Handshake C0+C1
    6   0.001732    127.0.0.1 → 127.0.0.1    RTMP 3139 Handshake S0+S1+S2
    8   0.001764    127.0.0.1 → 127.0.0.1    RTMP 1602 Handshake C2
   10   0.001789    127.0.0.1 → 127.0.0.1    RTMP 82 Set Chunk Size 4096
   12   0.001801    127.0.0.1 → 127.0.0.1    RTMP 252 connect('twinx')
   14   0.001963    127.0.0.1 → 127.0.0.1    RTMP 317 Acknowledgement 2500000|Set Peer Bandwidth 2500000,Dynamic|Set Chunk Size 8192|_result('NetConnection.Connect.Success')
   16   0.001997    127.0.0.1 → 127.0.0.1    RTMP 107 releaseStream('1234')
   18   0.002010    127.0.0.1 → 127.0.0.1    RTMP 103 FCPublish('1234')
   20   0.002018    127.0.0.1 → 127.0.0.1    RTMP 99 createStream()
   22   0.002078    127.0.0.1 → 127.0.0.1    RTMP 107 _result()
   24   0.002104    127.0.0.1 → 127.0.0.1    RTMP 112 publish('1234')
   26   0.002153    127.0.0.1 → 127.0.0.1    RTMP 201 Stream Begin 1|_result('NetStream.Publish.Start')
   53  11.747510    127.0.0.1 → 127.0.0.1    RTMP 105 FCUnpublish()
```

### (Healthy)


After [65926c904a1bd66a9d07d49e8d820ac012ab4676](https://github.com/kris-nova/twinx/commit/65926c904a1bd66a9d07d49e8d820ac012ab4676)

```
[nova@emily]: ~/twinx>$ tshark -r twinx.pcap  | grep -i RTMP | grep -v Data
    4   0.000138    127.0.0.1 → 127.0.0.1    RTMP 1603 Handshake C0+C1
    6   0.000295    127.0.0.1 → 127.0.0.1    RTMP 3139 Handshake S0+S1+S2
    8   0.000340    127.0.0.1 → 127.0.0.1    RTMP 1602 Handshake C2
   10   0.000361    127.0.0.1 → 127.0.0.1    RTMP 82 Set Chunk Size 4096
   12   0.000369    127.0.0.1 → 127.0.0.1    RTMP 252 connect('twinx')
   14   0.000469    127.0.0.1 → 127.0.0.1    RTMP 317 Acknowledgement 2500000|Set Peer Bandwidth 2500000,Dynamic|Set Chunk Size 8192|_result('NetConnection.Connect.Success')
   16   0.000523    127.0.0.1 → 127.0.0.1    RTMP 107 releaseStream('1234')
   18   0.000555    127.0.0.1 → 127.0.0.1    RTMP 103 FCPublish('1234')
   20   0.000577    127.0.0.1 → 127.0.0.1    RTMP 99 createStream()
   22   0.000736    127.0.0.1 → 127.0.0.1    RTMP 107 _result()
   24   0.000767    127.0.0.1 → 127.0.0.1    RTMP 112 publish('1234')
   26   0.000850    127.0.0.1 → 127.0.0.1    RTMP 202 Stream Begin 1|onStatus('NetStream.Publish.Start')
```


---

# Spec

See [Page 49](https://www.adobe.com/content/dam/acom/en/devnet/rtmp/pdf/rtmp_specification_1.0.pdf) for a working `Publish` transaction.