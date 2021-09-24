import BaseService from './base';
import Communicator from '../communicator';

import { FLAGS_EMPTY } from '../consts';
import { FLAP, SNAC, TLV } from '../structures';
import { word } from '../structures/bytes';

export default class BuddyListManagement extends BaseService {
  constructor(communicator : Communicator) {
    super({service: 0x03, version: 0x01}, communicator)
  }

  override handleMessage(message : FLAP) {
    if (!(message.payload instanceof SNAC)) {
      throw new Error('Expected SNACs')
    }

    if (message.payload.subtype === 0x02) {
      const resp = new FLAP(0x02, this.nextReqID,
        new SNAC(0x03, 0x03,  [
          new TLV(0x01, word(600)), // 600 max buddies
          new TLV(0x02, word(750)), // 750 max watchers
          new TLV(0x03, word(512)), // 512 max online notifications ?
        ]));
      
      this.send(resp);
      return;
    }
  }
}
