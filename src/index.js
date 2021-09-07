const net = require('net');
const { logDataStream } = require('./util');
const { FLAP } = require('./structures');

const server = net.createServer((socket) => {
  console.log('client connected...');
  socket.setTimeout(5000);

  socket.on('error', (e) => {
    console.error('socket encountered an error:', e);
  });

  socket.on('data', (data) => {
    const flap = FLAP.fromBuffer(Buffer.from(data, 'hex'));
    console.log('RECV', flap.toString());
  });

  socket.on('timeout', () => {
    console.log('socket timeout');
    socket.end();
  })

  socket.on('end', () => {
    console.log('client disconnected...');
  });

  const hello = new FLAP(0x01, 0, Buffer.from([0x00, 0x00, 0x00, 0x01]));
  socket.write(hello.toBuffer());

  /* 1. on connection, server sends

    2a      FLAP
    01      channel 1
    00 01   datagram #1
    00 04   4 bytes of data 
    00 00 00 1
  */

  /* 2. client responds
    2a      FLAP
    01      channel 1
    51 11   datagram 11
    00 04   4 bytes
    00 00 00 01
  */

  /* 3. server ACK */

  /* 4. client sends username
    2a            FLAP
    02            channel 2 (SNAC)
    51 12         datagram 12
    00 12         18 bytes of data
    00 17         Service (Authorization/registration service)
    00 06         Family  (Request md5 authkey)
    00 00         Flags
    00 00 00 00   SNAC request id
    00 01         TLV.Type(0x01) - screen name
    00 04         TLV.Length (4)
    74 6f 6f 66   toof
  */

    /* 5. server responds
      2a          FLAP
      02          Channel 2 (SNAC)
      00 02       datagram 2
      00 16       22 bytes
      00 17       Service (Authorization/registration service)
      00 07       Server md5 authkey response
                  This snac contain server generated auth key. Client should use it to crypt password.
      00 00       Flags
      00 00 00 00 SNAC request ID
      00 0a       Length (10 bytes)


    */


});

server.on('error', (err) => {
  throw err;
});

server.listen(5190, () => {
  console.log('OSCAR ready on :5190');
});
