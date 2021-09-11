import BaseService from './base';
import Communicator from '../communicator';
import { FLAP, Rate, RateClass, RatedServiceGroup, RateGroupPair, SNAC } from '../structures';
import { FLAGS_EMPTY } from '../consts';

export default class GenericServiceControls extends BaseService {
  constructor(communicator : Communicator) {
    super({service: 0x01, version: 0x03}, communicator)
  }

  override handleMessage(message : FLAP) {
    if (!(message.payload instanceof SNAC)) {
      throw new Error('Require SNAC');
    }

    if (message.payload.subtype === 0x06) { // Client ask server for rate limits info
      const resp = new FLAP(0x02, this._getNewSequenceNumber(),
        SNAC.forRateClass(0x01, 0x07, FLAGS_EMPTY, 0, [
          new Rate(
            new RateClass(1, 80, 2500, 2000, 1500, 800, 3400 /*fake*/, 6000, 0, 0),
            new RatedServiceGroup(1, [new RateGroupPair(0x00, 0x00)])
          )
        ]))
      this.send(resp);
      return;
    }

    if (message.payload.subtype === 0x0e) { // Client requests own online information
      console.log('should send back online presence info');
      return;
    }

    if (message.payload.subtype === 0x17) {
      const serviceVersions : Buffer[] = [];
      Object.values(this.communicator.services).forEach((subtype) => {
        serviceVersions.push(Buffer.from([0x00, subtype.service, 0x00, subtype.version]));
      });
      const resp = new FLAP(0x02, this._getNewSequenceNumber(),
        new SNAC(0x01, 0x18, FLAGS_EMPTY, 0, Buffer.concat(serviceVersions)));
      this.send(resp);
      return;
    }
  }
}
