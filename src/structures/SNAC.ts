import assert from "assert";
import { TLV } from "./TLV";

export class RateClass {
  constructor(
    public ID: number,
    public WindowSize: number,
    public ClearLevel: number,
    public AlertLevel: number,
    public LimitLevel: number,
    public DisconnectLevel: number,
    public CurrentLevel: number,
    public MaxLevel: number,
    public LastTime: number,
    public CurrentStat: number,
  ){}

  toBuffer() : Buffer {
    const buf = Buffer.alloc(35, 0x00);
    buf.writeUInt16BE(this.ID, 0);
    buf.writeUInt32BE(this.WindowSize, 2);
    buf.writeUInt32BE(this.ClearLevel, 6);
    buf.writeUInt32BE(this.AlertLevel, 10);
    buf.writeUInt32BE(this.LimitLevel, 14);
    buf.writeUInt32BE(this.DisconnectLevel,18);
    buf.writeUInt32BE(this.CurrentLevel, 22);
    buf.writeUInt32BE(this.MaxLevel, 26);
    buf.writeUInt32BE(this.LastTime, 30);
    buf.writeUInt8(this.CurrentStat, 34);
    return buf;
  }
}

export class RateGroupPair {
  constructor(public family : number, public service : number) {}
  toBuffer() : Buffer {
    const buf = Buffer.alloc(4, 0x00);
    buf.writeInt16BE(this.family, 0);
    buf.writeInt16BE(this.service, 2);
    return buf;
  }
}

export class RatedServiceGroup {
  constructor(public rateGroupID : number, public pairs : RateGroupPair[]){}

  toBuffer() : Buffer {
    const ratedServiceGroupHeader = Buffer.alloc(4, 0x00);
    ratedServiceGroupHeader.writeInt16BE(this.rateGroupID);
    ratedServiceGroupHeader.writeInt16BE(this.pairs.length, 2);
    const pairs = this.pairs.map((pair) => pair.toBuffer());
    return Buffer.concat([ratedServiceGroupHeader, ...pairs]);
  }
}

export class Rate {
  constructor(public rateClass : RateClass, public ratedServiceGroup : RatedServiceGroup) {}
  toBuffer() : Buffer {
    return Buffer.concat([this.rateClass.toBuffer(), this.ratedServiceGroup.toBuffer()]);
  }
}

export class SNAC {
  constructor(public family : number, public service : number, public flags : Buffer, public requestID : number , public payload : (TLV[] | Buffer) = Buffer.alloc(0)) {
    this.family = family;
    this.service = service;
    this.flags = flags;
    this.requestID = requestID;
    this.payload = payload;
  }

  static fromBuffer(buf : Buffer, payloadLength = 0) {
    assert(buf.length >= 10, 'Expected 10 bytes for SNAC header');
    const family = buf.slice(0,2).readInt16BE(0);
    const service = buf.slice(2,4).readInt16BE(0);
    const flags = buf.slice(4, 6);
    const requestID = buf.slice(6, 10).readInt32BE(0);
    let payload : Buffer | TLV[]; // SNACs can have multiple payload

    // Some SNACs don't have TLV payloads
    if (family === 0x01 && service === 0x17 ||
        family === 0x01 && service === 0x07 ||
        family === 0x01 && service === 0x08 ||
        family === 0x01 && service === 0x0e) {
      payload = buf.slice(10, 10 + payloadLength);
    } else {
      payload = [];
      // Try to parse TLVs
      let payloadIdx = 10;
      let cb = 0, cbLimit = 20; //circuit breaker
      while (payloadIdx < payloadLength && cb < cbLimit) {
        const tlv = TLV.fromBuffer(buf.slice(payloadIdx));
        payload.push(tlv);
        payloadIdx += tlv.length + 4; // 4 bytes for TLV type + payload length
        cb++;
      }
      if (cb === cbLimit) {
        console.error('Application error, cb limit reached');
        process.exit(1);
      }
    }

    return new SNAC(family, service, flags, requestID, payload);
  }

  static forRateClass(family : number, service : number, flags : Buffer, requestID : number, rates : Rate[]) : SNAC {
    const payloadHeader = Buffer.alloc(2, 0x00);
    payloadHeader.writeUInt16BE(rates.length);

    const payloadBody = rates.map((rateClass) => rateClass.toBuffer());
    const payload = Buffer.concat([payloadHeader, ...payloadBody]);

    return new SNAC(family, service, flags, requestID, payload);
  }

  toString() {
    return `SNAC(${this.family.toString(16)},${this.service.toString(16)}) #${this.requestID}\n  ${this.payload}`;
  }

  toBuffer() {
    const SNACHeader = Buffer.alloc(10, 0, 'hex');
    SNACHeader.writeUInt16BE(this.family);
    SNACHeader.writeUInt16BE(this.service, 2);
    SNACHeader.set(this.flags, 4);
    SNACHeader.writeUInt32BE(this.requestID, 6);

    let payload : Buffer[] = [];
    if (this.payload instanceof Buffer) {
      payload = [this.payload];
    } else if (this.payload.length && this.payload[0] instanceof TLV) {
      payload = (this.payload as TLV[]).map((thing : TLV) => thing.toBuffer());
    }

    return Buffer.concat([SNACHeader, ...payload]);
  }
}
