const crypto = require('crypto');
const BaseService = require('./base');
const { FLAP, SNAC, TLV } = require('../structures');

const { AIM_MD5_STRING, FLAGS_EMPTY } = require('../consts');

const users = {
  toof: 'foo',
};

class AuthorizationRegistrationService extends BaseService {
  constructor(communicator) {
    super({ family: 0x17, version: 0x01 }, communicator);
    this.cipher = "HARDY";
  }

  handleMessage(message) {
    switch (message.payload.service) {
      case 0x02: // Client login request (md5 login sequence)
        const tlvs = message.payload.tlvs;
        const clientNameTLV = tlvs.find((tlv) => tlv.type === 0x03);
        console.log("Attempting connection from", clientNameTLV.payload.toString('ascii'));

        const userTLV = tlvs.find((tlv) => tlv.type === 0x01);
        const username = userTLV.payload.toString('ascii');

        if (!users[username]) {
          const authResp = new FLAP(2, this._getNewSequenceNumber(),
          new SNAC(0x17, 0x03, FLAGS_EMPTY, 0, [
            new TLV(0x0001, Buffer.from(username)), // username
            new TLV(0x0008, Buffer.from([0x00, 0x04])) // incorrect nick/password
          ]));
          
          this.send(authResp);
          return;
        }

        const passwordHashTLV = tlvs.find((tlv) => tlv.type === 0x25);

        const pwHash = crypto.createHash('md5');
        pwHash.update(this.cipher);
        pwHash.update(users[username]);
        pwHash.update(AIM_MD5_STRING);
        const digest = pwHash.digest('hex');

        if (digest !== passwordHashTLV.payload.toString('hex')) {
          console.log('Invalid password for', username);
          const authResp = new FLAP(2, this._getNewSequenceNumber(),
          new SNAC(0x17, 0x03, FLAGS_EMPTY, 0, [
            new TLV(0x0001, Buffer.from(username)), // username
            new TLV(0x0008, Buffer.from([0x00, 0x04])) // incorrect nick/password
          ]));
          this.send(authResp);
          return;
        }

        const authResp = new FLAP(2, this._getNewSequenceNumber(),
        new SNAC(0x17, 0x03, FLAGS_EMPTY, 0, [
          new TLV(0x01, Buffer.from(username)), // username
          new TLV(0x05, Buffer.from('10.0.1.29:5190')), // BOS address
          new TLV(0x06, Buffer.from('im a cookie uwu')) // Authorization cookie
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
