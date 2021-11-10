import BaseService from './base';
import Communicator from '../communicator';
import { FLAP, SNAC, TLV } from '../structures';
import { word } from '../structures/bytes';

const emailToUin : {[key: string]: string[]} = { 
  'bob@example.com': ['bobX0X0'],
};

export default class UserLookup extends BaseService {
  constructor(communicator : Communicator) {
    super({service: 0x0a, version: 0x01}, [0x02], communicator)
  }

  override handleMessage(message: FLAP) {
    if (!(message.payload instanceof SNAC)) {
      throw new Error('Require SNAC');
    }

    // Search for a user by email address
    // TODO: don't return users that don't want to be found via email
    if (message.payload.subtype === 0x02) {
      if (!(message.payload instanceof Buffer)) {
        // 0x0e: Incorrect SNAC format
        const incorrectFormatResp = new FLAP(2, this.nextReqID, new SNAC(0x0a, 0x01, word(0x0e)));
        this.send(incorrectFormatResp);
        return;
      }

      const email = message.payload.payload.toString();
      if (!emailToUin[email]) {
        // 0x14: No Match
        const noResult = new FLAP(2, this.nextReqID, new SNAC(0x0a, 0x01, word(0x14)));
        this.send(noResult);
        return;
      }

      // Return list of TLVs of matching UINs
      const results = emailToUin[email].map((uin) => new TLV(0x01, Buffer.from(uin)))
      const resp = new FLAP(2, this.nextReqID, new SNAC(0x0a, 0x03, results));
      this.send(resp);
    }
  }
}
