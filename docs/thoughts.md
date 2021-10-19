# Scenario 1.

### Twinx Server with Publish and Play

Play Client (1) _VLC_
Publish Client (2) _OBS_
Server (3) _Twinx_


### Start playing (VLC)

Stream ID: `1234`

 - Handshake with server
 - Connect TX/RX
 - NetStream.Play.Start

### Start publishing (OBS)

Stream ID: `1234`

 - Handshake with the server
 - Connect TX/RX
 - NetStream.Publish.Start
 - StreamBegin() 
   - 