import assert from "assert";
import { FLAGS_EMPTY } from "../consts";
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
  constructor(public service : number, public subtype : number) {}
  toBuffer() : Buffer {
    const buf = Buffer.alloc(4, 0x00);
    buf.writeInt16BE(this.service, 0);
    buf.writeInt16BE(this.subtype, 2);
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

let snacID = 0x2000;

export class SNAC {
  constructor(public service : number, public subtype : number, public payload : (TLV[] | Buffer) = Buffer.alloc(0),   public requestID : number = 0, public flags : Buffer = FLAGS_EMPTY) {
    this.service = service;
    this.subtype = subtype;
    this.payload = payload;

    this.requestID = requestID || (snacID++);
    this.flags = flags;
  }

  static fromBuffer(buf : Buffer, payloadLength = 0) {
    assert(buf.length >= 10, 'Expected 10 bytes for SNAC header');
    const service = buf.slice(0,2).readInt16BE(0);
    const subtype = buf.slice(2,4).readInt16BE(0);
    const flags = buf.slice(4, 6);
    const requestID = buf.slice(6, 10).readInt32BE(0);
    let payload : Buffer | TLV[]; // SNACs can have multiple payload

    // Some SNACs don't have TLV payloads
    // Maybe this should be something that the service does itself when it
    // wants to respond to a message;
    if (service === 0x01 && subtype === 0x17 ||
        service === 0x01 && subtype === 0x14 ||
        service === 0x01 && subtype === 0x07 ||
        service === 0x01 && subtype === 0x08 ||
        service === 0x01 && subtype === 0x0e ||
        service === 0x04 && subtype === 0x02 ||
        service === 0x09 && subtype === 0x04 ||
        service === 0x0a && subtype === 0x02 ||
        service === 0x04 && subtype === 0x06) {
      payload = buf.slice(10, 10 + payloadLength);
    } else {
      payload = TLV.fromBufferBlob(buf.slice(10));
    }

    return new SNAC(service, subtype, payload, requestID, flags);
  }

  static forRateClass(service : number, subtype : number, rates : Rate[]) : SNAC {
    const payloadHeader = Buffer.alloc(2, 0x00);
    payloadHeader.writeUInt16BE(rates.length);

    const payloadBody = rates.map((rateClass) => rateClass.toBuffer());
    const payload = Buffer.concat([payloadHeader, ...payloadBody]);

    return new SNAC(service, subtype, payload);
  }

  toString() {
    return `SNAC(${this.service.toString(16)},${this.subtype.toString(16)}) #${this.requestID}\n  ${this.payload}`;
  }

  toBuffer() {
    const SNACHeader = Buffer.alloc(10, 0, 'hex');
    SNACHeader.writeUInt16BE(this.service);
    SNACHeader.writeUInt16BE(this.subtype, 2);
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
