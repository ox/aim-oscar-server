import assert from 'assert';

import { FLAP, SNAC, TLV } from '../src/structures';
import {FLAGS_EMPTY} from "../src/consts";

const tests = [
  () => {
    // Construct and test a CLI_AUTH_REQUEST
    const md5_auth_req = new FLAP(0x02, 0, new SNAC(0x17, 0x06,  [new TLV(0x01, Buffer.from("toof"))]));
    assert(md5_auth_req.channel === 2);
    assert(md5_auth_req.payload instanceof SNAC);
    assert(md5_auth_req.payload.service === 23);
    assert(md5_auth_req.payload.subtype === 6);
    assert(md5_auth_req.payload.payload.length === 1);
    assert.equal(md5_auth_req.payload.payload.length, 1);
    assert(md5_auth_req.payload.payload[0] instanceof TLV);
  },
  () => {
    // Test FLAP.length calculation and consuming multiple messages
    const dataStr = `
    2a 02 4b 11 00 0a 00 01
    00 0e 00 00 00 00 00 00
    2a 02 4b 12 00 0a 00 02
    00 02 00 00 00 00 00 00
    `.trim().replace(/\s+/g, '');
    let data = Buffer.from(dataStr, 'hex');

    const message = FLAP.fromBuffer(data);
    assert(message.channel === 2);
    assert(message.payloadLength === 10);
    assert(message.length === 16);
    assert(message.payload instanceof SNAC);
    assert((message.payload as SNAC).service === 1);
    assert((message.payload as SNAC).subtype === 0x0e);

    data = data.slice(message.length);
    const secondMessage = FLAP.fromBuffer(data);
    assert(secondMessage.length === 16);
    assert(secondMessage.payload instanceof SNAC);
    assert((secondMessage.payload as SNAC).service === 2);
    assert((secondMessage.payload as SNAC).subtype === 0x02);
  }
];

tests.forEach((testFn) => {
  testFn();
});
