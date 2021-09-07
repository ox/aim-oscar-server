const net = require('net');
const assert = require('assert');

function chunkString(str, len) {
  const size = Math.ceil(str.length/len)
  const r = Array(size)
  let offset = 0
  
  for (let i = 0; i < size; i++) {
    r[i] = str.substr(offset, len)
    offset += len
  }
  
  return r
}

function logDataStream(data){  
  const strs = chunkString(data.toString('hex'), 16);
  return strs.map((str) => chunkString(str, 2).join(' ')).join('\n');
}

function TLV(buf) {
  this.type = buf.slice(0, 2);
  this.len = parseInt(buf.slice(2, 4).toString('hex'), 16);
  this.payload = buf.slice(4, 4 + this.len);
  this.toString = () => `TLV(${this.type.toString('hex')}, ${this.len}, ${this.payload.toString('ascii')})`;
}

function SNAC(buf) {
  this.family = buf.slice(0,2).readInt16BE(0);
  this.service = buf.slice(2,4).readInt16BE(0);
  this.flags = buf.slice(4, 6);
  this.requestID = parseInt(buf.slice(6, 10).toString('hex'), 16);
  this.payload = new TLV(buf.slice(10));
  this.toString = () => `SNAC(${this.family.toString(16)},${this.service.toString(16)}) #${this.requestID}\n  ${this.payload}`;
}

function FLAP(buf) {
  assert.equal(buf[0], 0x2a, 'Expected 0x2a FLAP header');
  this.channel = buf[1];
  this.datagramNumber = parseInt(buf.slice(2,4).toString('hex'), 16);
  this.payloadLength = parseInt(buf.slice(4, 6).toString('hex'), 16);
  this.payload = buf.slice(6, 6 + this.payloadLength);
  this.toString = () => `ch:${this.channel}, dn: ${this.datagramNumber}, len: ${this.payloadLength}, payload:\n  ${ this.payload instanceof SNAC ? this.payload.toString() : logDataStream(this.payload).split('\n').join('\n  ')}`;

  if (this.channel === 2) {
    this.payload = new SNAC(this.payload);
  }
}

const server = net.createServer((socket) => {
  console.log('client connected...');
  socket.setTimeout(5000);

  socket.on('error', (e) => {
    console.error('socket encountered an error:', e);
  });

  socket.on('data', (data) => {
    const flap = new FLAP(Buffer.from(data, 'hex'));
    console.log('RECV', flap.toString());
  });

  socket.on('timeout', () => {
    console.log('socket timeout');
    socket.end();
  })

  socket.on('end', () => {
    console.log('client disconnected...');
  });

  const hello = Buffer.from(new Uint8Array([0x2a, 0x01, 0, 0x01, 0, 0x04, 0x00, 0x00, 0x00, 0x01]));
  socket.write(hello);

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

server.listen(5190, '10.0.1.29', () => {
  console.log('OSCAR ready on 10.0.1.29:5190');
});
