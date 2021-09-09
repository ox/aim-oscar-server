import assert from "assert"

import { SNAC } from "./SNAC";
import { logDataStream } from '../util';

export class FLAP {
  static fromBuffer(buf : Buffer) {
    assert.equal(buf[0], 0x2a, 'Expected 0x2a at start of FLAP header');
    assert(buf.length >= 6, 'Expected at least 6 bytes for FLAP header');
    const channel = buf.readInt8(1);
    const sequenceNumber = buf.slice(2,4).readInt16BE(0);
    const payloadLength = buf.slice(4, 6).readInt16BE(0);
    let payload : Buffer | SNAC = buf.slice(6, 6 + payloadLength);

    if (channel === 2) {
      payload = SNAC.fromBuffer(payload, payloadLength);
    }

    return new FLAP(channel, sequenceNumber, payload)
  }

  payloadLength: number;

  constructor(public channel: number, public sequenceNumber: number, public payload: Buffer | SNAC) {
    this.channel = channel;
    this.sequenceNumber = sequenceNumber;

    this.payload = payload;

    if (payload instanceof SNAC) {
      this.payloadLength = payload.toBuffer().length;
    } else {
      this.payloadLength = payload.length;
    }
  }

  toString() {
    let payload = this.payload.toString();
    if (this.payload instanceof Buffer) {
      payload = logDataStream(this.payload).split('\n').join('\n  ');
    }
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
