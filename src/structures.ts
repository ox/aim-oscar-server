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
    return `TLV(0x${this.type.toString(16).padStart(2, '0')}, ${this.len}, ${this.payload.toString('ascii')})`;
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
    assert(buf.length >= 10, 'Expected 10 bytes for SNAC header');
    const family = buf.slice(0,2).readInt16BE(0);
    const service = buf.slice(2,4).readInt16BE(0);
    const flags = buf.slice(4, 6);
    const requestID = buf.slice(6, 10).readInt32BE(0);
    const tlvs = []; // SNACs can have multiple TLVs

    let tlvsIdx = 10;
    let cb = 0, cbLimit = 20; //circuit breaker
    while (tlvsIdx < payloadLength && cb < cbLimit) {
      const tlv = TLV.fromBuffer(buf.slice(tlvsIdx));
      tlvs.push(tlv);
      tlvsIdx += tlv.len + 4; // 4 bytes for TLV type + tlvs length
      cb++;
    }
    if (cb === cbLimit) {
      console.error('Application error, cb limit reached');
      process.exit(1);
    }
    
    return new SNAC(family, service, flags, requestID, tlvs);
  }

  constructor(family, service, flags, requestID, tlvs = []) {
    this.family = family;
    this.service = service;
    this.flags = flags;
    this.requestID = requestID;
    this.tlvs = tlvs;
  }

  toString() {
    return `SNAC(${this.family.toString(16)},${this.service.toString(16)}) #${this.requestID}\n  ${this.tlvs}`;
  }

  toBuffer() {
    const SNACHeader = Buffer.alloc(10, 0, 'hex');
    SNACHeader.writeUInt16BE(this.family);
    SNACHeader.writeUInt16BE(this.service, 2);
    SNACHeader.writeUInt16BE(this.flags, 4);
    SNACHeader.writeUInt32BE(this.requestID, 6);

    const payload = this.tlvs.map((thing) => {
      if (thing instanceof TLV) {
        return thing.toBuffer();
      }
      return thing;
    });

    return Buffer.concat([SNACHeader, ...payload]);
  }
}

class FLAP {
  static fromBuffer(buf) {
    assert.equal(buf[0], 0x2a, 'Expected 0x2a at start of FLAP header');
    assert(buf.length >= 6, 'Expected at least 6 bytes for FLAP header');
    const channel = buf.readInt8(1);
    const sequenceNumber = buf.slice(2,4).readInt16BE(0);
    const payloadLength = buf.slice(4, 6).readInt16BE(0);
    let payload = buf.slice(6, 6 + payloadLength);

    if (channel === 2) {
      payload = SNAC.fromBuffer(payload, payloadLength);
    }

    return new FLAP(channel, sequenceNumber, payload)
  }

  constructor(channel, sequenceNumber, payload) {
    this.channel = channel;
    this.sequenceNumber = sequenceNumber;

    this.payload = payload;
    this.payloadLength = this.payload.length;

    if (payload instanceof SNAC) {
      this.payloadLength = payload.toBuffer().length;
    }
  }

  toString() {
    const hasSnac = this.payload instanceof SNAC;
    const payload = hasSnac ? this.payload.toString() : logDataStream(this.payload).split('\n').join('\n  ');
    return `ch:${this.channel}, dn: ${this.sequenceNumber}, len: ${this.payloadLength}, payload:\n  ${payload}`
  }

  toBuffer() {
    const FLAPHeader = Buffer.alloc(6, 0, 'hex');
    FLAPHeader.writeInt8(0x2a, 0);
    FLAPHeader.writeInt8(this.channel, 1);
    FLAPHeader.writeInt16BE(this.sequenceNumber, 2);
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
