import BaseService from './base';
import Communicator from '../communicator';

import { FLAGS_EMPTY } from '../consts';
import { FLAP, SNAC, TLV } from '../structures';
import { word } from '../structures/bytes';


export default class PrivacyManagement extends BaseService {
  private permissionMask: number = 0xffff; // everyone

  constructor(communicator : Communicator) {
    super({service: 0x09, version: 0x01}, communicator)
  }

  override handleMessage(message : FLAP) {
    if (!(message.payload instanceof SNAC)) {
      throw new Error('Expected SNACs')
    }

    if (message.payload.subtype === 0x02) {
      const resp = new FLAP(0x02, this._getNewSequenceNumber(),
        new SNAC(0x09, 0x03, FLAGS_EMPTY, 0, [
          new TLV(0x01, word(200)), // max visible list size
          new TLV(0x02, word(200))  // max invisible list size
        ]));
      this.send(resp);
      return;
    }

    if (message.payload.subtype === 0x04) {
      // Client sends permission mask for classes of users that can talk to the client
      this.permissionMask = (message.payload.payload as Buffer).readUInt32BE();
      console.log('set permission mask', this.permissionMask.toString(16));
      return;
    }
  }
}
