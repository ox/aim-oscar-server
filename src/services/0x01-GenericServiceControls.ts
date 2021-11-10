import BaseService from './base';
import Communicator from '../communicator';
import { FLAP, Rate, RateClass, RatedServiceGroup, RateGroupPair, SNAC, TLV } from '../structures';
import {  USER_STATUS_VARIOUS, USER_STATUS } from '../consts';
import { char, word, dword, dot2num } from '../structures/bytes';

export default class GenericServiceControls extends BaseService {
  private allowViewIdle = false;
  private allowViewMemberSince = false;

  constructor(communicator : Communicator) {
    super({service: 0x01, version: 0x03}, [0x06, 0x0e, 0x14, 0x17], communicator)
  }

  override handleMessage(message : FLAP) {
    if (!(message.payload instanceof SNAC)) {
      throw new Error('Require SNAC');
    }

    if (message.payload.subtype === 0x06) { // Client ask server for rate limits info

      // HACK: set rate limits for all services. I can't tell which message subtypes they support so
      // make it set rate limits for everything under 0x21.
      const pairs : RateGroupPair[] = [];
      Object.values(this.communicator.services).forEach((service) => {
        // for (let subtype of service.supportedSubtypes) {
        for (let subtype = 0; subtype < 0x21; subtype++) {
          pairs.push(new RateGroupPair(service.service, subtype));
        }
      });

      const resp = new FLAP(0x02, this.nextReqID,
        SNAC.forRateClass(0x01, 0x07, [
          new Rate(
            new RateClass(1, 80, 2500, 2000, 1500, 800, 3400 /*fake*/, 6000, 0, 0),
            new RatedServiceGroup(1, pairs),
          )
        ]));
      this.send(resp);

      const motd = new FLAP(0x02, this.nextReqID,
        new SNAC(0x01, 0x13,  Buffer.concat([
          word(0x0004),
          (new TLV(0x0B, Buffer.from("Hello world!"))).toBuffer(),
        ])));
      this.send(motd);
      return;
    }

    if (message.payload.subtype === 0x0e) { // Client requests own online information
      const uin = this.communicator.user?.username || 'user';
      const warning = 0;
      const since = +(new Date('December 17, 1998 03:24:00'));
      const externalIP = dot2num(this.communicator.socket.remoteAddress!.split(':').pop()!);

      const tlvs : TLV[] = [
        new TLV(0x01, char(0x80)),
        new TLV(0x06, dword(USER_STATUS_VARIOUS.WEBAWARE | USER_STATUS_VARIOUS.DCDISABLED << 2 + USER_STATUS.ONLINE)),
        new TLV(0x0A, dword(externalIP)),
        new TLV(0x0F, dword(0)), // TODO: track idle time,
        new TLV(0x03, dword(Math.floor(Date.now() / 1000))),
        new TLV(0x1E, dword(0)), // Unknown
        new TLV(0x05, dword(Math.floor(since / 1000))),
        new TLV(0x0C, Buffer.concat([
          dword(externalIP),
          dword(5700), // DC TCP Port
          dword(0x04000000), // DC Type,
          word(0x0400), // DC Protocol Version
          dword(0), // DC Auth Cookie
          dword(0), // Web Front port
          dword(0x300), // Client Features ?
          dword(0), // Last Info Update Time
          dword(0), // last EXT info update time,
          dword(0), // last ext status update time
        ]))
      ];

      const payloadHeader = Buffer.alloc(1 + uin.length + 2 + 2);
      payloadHeader.writeInt8(uin.length);
      payloadHeader.set(Buffer.from(uin), 1);
      payloadHeader.writeInt16BE(warning, 1 + uin.length);
      payloadHeader.writeInt16BE(tlvs.length, 1 + uin.length + 2);

      const buf = Buffer.concat([payloadHeader, ...tlvs.map((tlv) => tlv.toBuffer())])

      const resp = new FLAP(0x02, this.nextReqID,
        new SNAC(0x01, 0x0f, buf));

      this.send(resp);
      return;
    }

    if (message.payload.subtype === 0x14) {
      /*
        Client setting privacy settings
          Bit 1 - Allows other AIM users to see how long you've been idle.
          Bit 2 - Allows other AIM users to see how long you've been a member.
      */
     const mask = (message.payload.payload as Buffer).readUInt32BE();
     this.allowViewIdle = (mask & 0x01) > 0;
     this.allowViewMemberSince = (mask & 0x02) > 0;
     console.log('allowViewIdle:', this.allowViewIdle, 'allowViewMemberSince', this.allowViewMemberSince);
     return;
    }

    if (message.payload.subtype === 0x17) {
      const serviceVersions : Buffer[] = [];
      Object.values(this.communicator.services).forEach((subtype) => {
        serviceVersions.push(Buffer.from([0x00, subtype.service, 0x00, subtype.version]));
      });
      const resp = new FLAP(0x02, this.nextReqID,
        new SNAC(0x01, 0x18,  Buffer.concat(serviceVersions)));
      this.send(resp);
      return;
    }
  }
}
