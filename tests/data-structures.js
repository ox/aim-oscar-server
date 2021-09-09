import assert from 'assert';

const { FLAP, SNAC, TLV } = require('../src/structures');

const tests = [
  () => {
    // Construct and test a CLI_AUTH_REQUEST
    const md5_auth_req = new FLAP(0x02, 0, new SNAC(0x17, 0x06, 0x0000, 0, [new TLV(0x0001, Buffer.from("toof"))]));
    assert(md5_auth_req.channel === 2);
    assert(md5_auth_req.payload instanceof SNAC);
    assert(md5_auth_req.payload.family === 23);
    assert(md5_auth_req.payload.service === 6);
    assert(md5_auth_req.payload.payload.length === 1);
    assert(md5_auth_req.payload.payload[0].len === 4);
  }
];

tests.forEach((testFn) => {
  testFn();
});
