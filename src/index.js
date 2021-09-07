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

class TLV {
  static fromBuffer(buf) {
    const type = buf.slice(0, 2).readInt16BE(0);
    const len = buf.slice(2, 4).readInt16BE(0)
    const payload = buf.slice(4, 4 + len);

    return new TLV(type, payload);
  }

  constructor(type, payload) {
    this.type = type;
    this.len = payload.length;
    this.payload = payload;
  }

  toString() {
    return `TLV(${this.type}, ${this.len}, ${this.payload.toString('ascii')})`;
  }
}

class SNAC {
  static fromBuffer(buf, payloadLength = 0) {
    const family = buf.slice(0,2).readInt16BE(0);
    const service = buf.slice(2,4).readInt16BE(0);
    const flags = buf.slice(4, 6);
    const requestID = buf.slice(6, 10).readInt32BE(0);
    const payload = []; // SNACs can have multiple TLVs

    let payloadIdx = 10;
    let cb = 0, cbLimit = 10; //circuit breaker
    while (payloadIdx < payloadLength && cb < cbLimit) {
      const tlv = TLV.fromBuffer(buf.slice(payloadIdx));
      payload.push(tlv);
      payloadIdx += tlv.len + 4; // 4 bytes for TLV type + payload length
      cb++;
    }
    if (cb === cbLimit) {
      console.error('Application error, cb limit reached');
      process.exit(1);
    }
    
    return new SNAC(family, service, flags, requestID, payload);
  }

  constructor(family, service, flags, requestID, payload) {
    this.family = family;
    this.service = service;
    this.flags = flags;
    this.requestID = requestID;
    this.payload = payload;
  }

  toString() {
    return `SNAC(${this.family.toString(16)},${this.service.toString(16)}) #${this.requestID}\n  ${this.payload}`;
  }
}

class FLAP {
  static fromBuffer(buf) {
    assert.equal(buf[0], 0x2a, 'Expected 0x2a FLAP header');
    const channel = parseInt(buf[1], 16);
    const datagramNumber = buf.slice(2,4).readInt16BE(0);
    const payloadLength = buf.slice(4, 6).readInt16BE(0);
    const payload = buf.slice(6, 6 + payloadLength);

    return new FLAP(channel, datagramNumber, payload)
  }

  constructor(channel, datagramNumber, payload) {
    this.channel = channel;
    this.datagramNumber = datagramNumber;
    this.payload = payload;
    this.payloadLength = this.payload.length;

    if (channel === 2) {
      this.payload = SNAC.fromBuffer(this.payload, this.payloadLength);
    }
  }

  toString() {
    const hasSnac = this.payload instanceof SNAC;
    const payload = hasSnac ? this.payload.toString() : logDataStream(this.payload).split('\n').join('\n  ');
    return `ch:${this.channel}, dn: ${this.datagramNumber}, len: ${this.payloadLength}, payload:\n  ${payload}`
  }
}

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

server.listen(5190, () => {
  console.log('OSCAR ready on :5190');
});
