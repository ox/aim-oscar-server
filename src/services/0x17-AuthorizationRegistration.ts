import crypto from 'crypto';
import BaseService from './base';
import Communicator, { User } from '../communicator';
import { FLAP, SNAC, TLV, ErrorCode, TLVType } from '../structures';
import { word } from '../structures/bytes';

const { AIM_MD5_STRING, FLAGS_EMPTY } = require('../consts');

const users : {[key: string]: User} = {
  'toof': {
    uin: '156089',
    username: 'toof',
    password: 'foo',
    memberSince: new Date('December 17, 1998 03:24:00'),
  }
};

export default class AuthorizationRegistrationService extends BaseService {
  private cipher : string;

  constructor(communicator : Communicator) {
    super({ service: 0x17, version: 0x01 }, communicator);
    this.cipher = "HARDY";
  }

  override handleMessage(message : FLAP) {
    if (message.payload instanceof Buffer) {
      console.log('Wont handle Buffer payload');
      return;
    }

    switch (message.payload.subtype) {
      case 0x02: // Client login request (md5 login sequence)
        const payload = message.payload.payload;
        const clientNameTLV = payload.find((tlv) => tlv instanceof TLV && tlv.type === TLVType.ClientName);
        if (!clientNameTLV || !(clientNameTLV instanceof TLV)) {
          return;
        }
        console.log("Attempting connection from", clientNameTLV.payload.toString('ascii'));

        const userTLV = payload.find((tlv) => tlv instanceof TLV && tlv.type === TLVType.User);
        if (!userTLV  || !(userTLV instanceof TLV)) {
          return;
        }

        const username = userTLV.payload.toString('ascii');

        if (!users[username]) {
          const authResp = new FLAP(2, this.nextReqID,
          new SNAC(0x17, 0x03,  [
            TLV.forUsername(username), // username
            TLV.forError(ErrorCode.IncorrectNick) // incorrect nick/password
          ]));
          
          this.send(authResp);
          return;
        }

        const passwordHashTLV = payload.find((tlv) => tlv instanceof TLV && tlv.type === TLVType.PasswordHash);
        if (!passwordHashTLV || !(passwordHashTLV instanceof TLV)) {
          return;
        }

        const pwHash = crypto.createHash('md5');
        pwHash.update(this.cipher);
        pwHash.update(users[username].password);
        pwHash.update(AIM_MD5_STRING);
        const digest = pwHash.digest('hex');

        if (digest !== (passwordHashTLV as TLV).payload.toString('hex')) {
          console.log('Invalid password for', username);
          const authResp = new FLAP(2, this.nextReqID,
          new SNAC(0x17, 0x03,  [
            TLV.forUsername(username), // username
            TLV.forError(ErrorCode.IncorrectNick) // incorrect nick/password
          ]));
          this.send(authResp);

          // Close this connection
          const plsLeave = new FLAP(4, this.nextReqID, Buffer.from([]));
          this.send(plsLeave);
          return;
        }

        const chatHost = this.communicator.socket.localAddress.split(':').pop() + ':5191';
      
        const authResp = new FLAP(2, this.nextReqID,
        new SNAC(0x17, 0x03,  [
          TLV.forUsername(username), // username
          TLV.forBOSAddress(chatHost), // BOS address
          TLV.forCookie(JSON.stringify({cookie: 'uwu', user: users[username]})) // Authorization cookie
        ]));

        this.communicator.user = Object.assign({username}, users[username]);
        console.log(this.communicator.user);

        this.send(authResp);

        // tell them to leave
        const disconnectResp = new FLAP(4, this.nextReqID, Buffer.alloc(0));
        this.send(disconnectResp);

        return;
      case 0x06: // Request md5 authkey
        const MD5AuthKeyHeader = Buffer.alloc(2, 0xFF, 'hex');
        MD5AuthKeyHeader.writeUInt16BE(this.cipher.length);
        const md5ReqResp = new FLAP(2, this.nextReqID,
          new SNAC(0x17, 0x07, 
            Buffer.concat([MD5AuthKeyHeader, Buffer.from(this.cipher, 'binary')]),
          ));
        this.send(md5ReqResp);
        break;
    }
  }
}

module.exports = AuthorizationRegistrationService;
