import assert from "assert";
import { TLV } from "./TLV";

export class SNAC {
  static fromBuffer(buf : Buffer, payloadLength = 0) {
    assert(buf.length >= 10, 'Expected 10 bytes for SNAC header');
    const family = buf.slice(0,2).readInt16BE(0);
    const service = buf.slice(2,4).readInt16BE(0);
    const flags = buf.slice(4, 6);
    const requestID = buf.slice(6, 10).readInt32BE(0);
    const tlvs : TLV[] = []; // SNACs can have multiple TLVs

    let tlvsIdx = 10;
    let cb = 0, cbLimit = 20; //circuit breaker
    while (tlvsIdx < payloadLength && cb < cbLimit) {
      const tlv = TLV.fromBuffer(buf.slice(tlvsIdx));
      tlvs.push(tlv);
      tlvsIdx += tlv.length + 4; // 4 bytes for TLV type + tlvs length
      cb++;
    }
    if (cb === cbLimit) {
      console.error('Application error, cb limit reached');
      process.exit(1);
    }
    
    return new SNAC(family, service, flags, requestID, tlvs);
  }

  constructor(public family : number, public service : number, public flags : Buffer, public requestID : number , public tlvs : Array<TLV | Buffer> = []) {
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
    SNACHeader.set(this.flags, 4);
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
