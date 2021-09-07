const assert = require('assert');

const { logDataStream } = require('./util');

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

  toBuffer() {
    const TLVHeader = Buffer.alloc(4, 0, 'hex');
    TLVHeader.writeUInt16BE(this.type);
    TLVHeader.writeUInt16BE(this.len, 2);
    return Buffer.concat([TLVHeader, this.payload]);
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

  toBuffer() {
    const SNACHeader = Buffer.alloc(10, 0, 'hex');
    SNACHeader.writeUInt16BE(this.family);
    SNACHeader.writeUInt16BE(this.service, 2);
    SNACHeader.writeUInt16BE(this.flags, 4);
    SNACHeader.writeUInt32BE(this.requestID, 6);

    const payload = this.payload.map((tlv) => tlv.toBuffer());
    return Buffer.concat([SNACHeader, ...payload]);
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

    if (payload instanceof SNAC) {
      this.payloadLength = payload.toBuffer().length;
    }

    if (channel === 2 && !(payload instanceof SNAC)) {
      this.payload = SNAC.fromBuffer(this.payload, this.payloadLength);
    }
  }

  toString() {
    const hasSnac = this.payload instanceof SNAC;
    const payload = hasSnac ? this.payload.toString() : logDataStream(this.payload).split('\n').join('\n  ');
    return `ch:${this.channel}, dn: ${this.datagramNumber}, len: ${this.payloadLength}, payload:\n  ${payload}`
  }

  toBuffer() {
    const FLAPHeader = Buffer.alloc(6, 0, 'hex');
    FLAPHeader.writeInt8(0x2a, 0);
    FLAPHeader.writeInt8(this.channel, 1);
    FLAPHeader.writeInt16BE(this.datagramNumber, 2);
    FLAPHeader.writeInt16BE(this.payloadLength, 4);

    let payload = this.payload;
    if (payload instanceof SNAC) {
      payload = payload.toBuffer();
    }

    return Buffer.concat([FLAPHeader, payload]);
  }
}

module.exports = {
  TLV,
  SNAC,
  FLAP,
};
