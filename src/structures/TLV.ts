import {ErrorCode} from "./ErrorCode";

import { USER_STATUS_VARIOUS, USER_STATUS } from "../consts";
import { char, word, dword } from './bytes';

export const enum TLVType {
  User = 0x01,
  ClientName = 0x03,
  GetServices = 0x06,
  PasswordHash = 0x25,
}

export class TLV {
  get length() : number {
    return this.payload.length;
  }

  static fromBuffer(buf : Buffer) {
    const type = buf.slice(0, 2).readInt16BE(0) as TLVType;
    const len = buf.slice(2, 4).readInt16BE(0)
    const payload = buf.slice(4, 4 + len);

    return new TLV(type, payload);
  }

  static forUsername(username : string) : TLV {
    return new TLV(0x01, Buffer.from(username));
  }

  static forBOSAddress(address : string ) : TLV {
    return new TLV(0x05, Buffer.from(address));
  }

  static forCookie(cookie : string) : TLV {
    return new TLV(0x06, Buffer.from(cookie));
  }

  static forError(errorCode : ErrorCode) : TLV {
    return new TLV(0x08, Buffer.from([0x00, errorCode]));
  }

  static forStatus(various : USER_STATUS_VARIOUS, status: USER_STATUS) {
    const varbuf = Buffer.alloc(2, 0x00);
    varbuf.writeUInt16BE(various);
    const statbuf = Buffer.alloc(2, 0x00);
    statbuf.writeUInt16BE(status);
    return new TLV(0x06, Buffer.concat([varbuf, statbuf]));
  }

  constructor(public type : TLVType, public payload : Buffer) {
    this.type = type;
    this.payload = payload;
  }

  toString() {
    return `TLV(0x${this.type.toString(16).padStart(2, '0')}, ${this.length}, ${this.payload.toString('ascii')})`;
  }

  toBuffer() {
    return Buffer.concat([
      word(this.type),
      word(this.length),
      this.payload,
    ]);
  }
}
