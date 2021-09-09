import crypto from 'crypto';
import BaseService from './base';
import Communicator from '../communicator';
import { FLAP, SNAC, TLV, ErrorCode, TLVType } from '../structures';

const { AIM_MD5_STRING, FLAGS_EMPTY } = require('../consts');

const users : {[key: string]: string} = {
  'toof': 'foo',
};

export default class AuthorizationRegistrationService extends BaseService {
  private cipher : string;

  constructor(communicator : Communicator) {
    super({ family: 0x17, version: 0x01 }, communicator);
    this.cipher = "HARDY";
  }

  override handleMessage(message : FLAP) {
    if (message.payload instanceof Buffer) {
      console.log('Wont handle Buffer payload');
      return;
    }

    switch (message.payload.service) {
      case 0x02: // Client login request (md5 login sequence)
        const tlvs = message.payload.tlvs;
        const clientNameTLV = tlvs.find((tlv) => tlv instanceof TLV && tlv.type === TLVType.ClientName);
        if (!clientNameTLV || !(clientNameTLV instanceof TLV)) {
          return;
        }
        console.log("Attempting connection from", clientNameTLV.payload.toString('ascii'));

        const userTLV = tlvs.find((tlv) => tlv instanceof TLV && tlv.type === TLVType.User);
        if (!userTLV  || !(userTLV instanceof TLV)) {
          return;
        }

        const username = userTLV.payload.toString('ascii');

        if (!users[username]) {
          const authResp = new FLAP(2, this._getNewSequenceNumber(),
          new SNAC(0x17, 0x03, FLAGS_EMPTY, 0, [
            TLV.forUsername(username), // username
            TLV.forError(ErrorCode.IncorrectNick) // incorrect nick/password
          ]));
          
          this.send(authResp);
          return;
        }

        const passwordHashTLV = tlvs.find((tlv) => tlv instanceof TLV && tlv.type === TLVType.PasswordHash);
        if (!passwordHashTLV || !(passwordHashTLV instanceof TLV)) {
          return;
        }

        const pwHash = crypto.createHash('md5');
        pwHash.update(this.cipher);
        pwHash.update(users[username]);
        pwHash.update(AIM_MD5_STRING);
        const digest = pwHash.digest('hex');

        if (digest !== (passwordHashTLV as TLV).payload.toString('hex')) {
          console.log('Invalid password for', username);
          const authResp = new FLAP(2, this._getNewSequenceNumber(),
          new SNAC(0x17, 0x03, FLAGS_EMPTY, 0, [
            TLV.forUsername(username), // username
            TLV.forError(ErrorCode.IncorrectNick) // incorrect nick/password
          ]));
          this.send(authResp);
          return;
        }

        const authResp = new FLAP(2, this._getNewSequenceNumber(),
        new SNAC(0x17, 0x03, FLAGS_EMPTY, 0, [
          TLV.forUsername(username), // username
          TLV.forBOSAddress('10.0.1.29:5190'), // BOS address
          TLV.forCookie('im a cookie uwu') // Authorization cookie
        ]));

        this.send(authResp);
        return;
      case 0x06: // Request md5 authkey
        const payload = Buffer.alloc(2, 0xFF, 'hex');
        payload.writeUInt16BE(this.cipher.length);
        const md5ReqResp = new FLAP(2, this._getNewSequenceNumber(),
          new SNAC(0x17, 0x07, FLAGS_EMPTY, 0, [
            Buffer.concat([payload, Buffer.from(this.cipher, 'binary')]),
          ]));
        this.send(md5ReqResp);
        break;
    }
  }
}

module.exports = AuthorizationRegistrationService;
