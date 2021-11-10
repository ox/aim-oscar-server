import BaseService from './base';
import Communicator from '../communicator';

import { FLAP, SNAC, TLV } from '../structures';
import { char, word, dword, dot2num } from '../structures/bytes';
import {  USER_STATUS, USER_STATUS_VARIOUS } from '../consts';

export default class LocationServices extends BaseService {
  constructor(communicator : Communicator) {
    super({service: 0x02, version: 0x01}, [0x02, 0x04], communicator)
  }

  override handleMessage(message : FLAP) {
    if (!(message.payload instanceof SNAC)) {
      throw new Error('Expecting SNACs for LocationServices')
    }

    // request location service parameters and limitations
    if (message.payload.subtype === 0x02) {
      const resp = new FLAP(0x02, this.nextReqID,
        new SNAC(0x02,0x03,  [
          new TLV(0x01, word(0x400)), // max profile length
          new TLV(0x02, word(0x10)),  // max capabilities
          new TLV(0x03, word(0xA)),   // unknown
          new TLV(0x04, word(0x1000)),
        ]));
      this.send(resp);
      return;
    }

    if (message.payload.subtype === 0x04) {
      // Client use this snac to set its location information (like client
      // profile string, client directory profile string, client capabilities).
      return;
    }
  }
}
