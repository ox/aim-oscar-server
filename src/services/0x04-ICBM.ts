import BaseService from './base';
import Communicator from '../communicator';

import { FLAGS_EMPTY } from '../consts';
import { FLAP, SNAC, TLV } from '../structures';
import { dword } from '../structures/bytes';

interface ChannelSettings {
  channel: number,
  messageFlags: number,
  maxMessageSnacSize: number,
  maxSenderWarningLevel: number,
  maxReceiverWarningLevel: number,
  minimumMessageInterval: number,
  unknown : number
}

export default class ICBM extends BaseService {
  private channel : ChannelSettings = {
    channel: 0,
    messageFlags: 3,
    maxMessageSnacSize: 512,
    maxSenderWarningLevel: 999,
    maxReceiverWarningLevel: 999,
    minimumMessageInterval: 0,
    unknown: 1000,
  };

  private channels : ChannelSettings[] = [];

  constructor(communicator : Communicator) {
    super({service: 0x04, version: 0x01}, communicator)
  }

  override handleMessage(message : FLAP) {
    if (!(message.payload instanceof SNAC)) {
      throw new Error('Expected SNACs')
    }

    if (message.payload.subtype === 0x02) {
      // client is telling us about it's ICBM capabilities (whatever)
      /*
        xx xx	 	word	 	channel to setup
        xx xx xx xx	 	dword	 	message flags
        xx xx	 	word	 	max message snac size
        xx xx	 	word	 	max sender warning level
        xx xx	 	word	 	max receiver warning level
        xx xx	 	word	 	minimum message interval (sec)
        00 00	 	word	 	unknown parameter (also seen 03 E8)
      */

      if (!(message.payload.payload instanceof Buffer)) {
        throw new Error('Expected Buffer payload for this SNAC');
      }

      const payload = message.payload.payload;
      const channel = payload.readUInt16BE(0);

      // TODO: set settings based on channel provided

      this.channel = {
        channel,
        messageFlags: payload.readUInt32BE(2),
        maxMessageSnacSize: payload.readUInt16BE(6),
        maxSenderWarningLevel: payload.readUInt16BE(8),
        maxReceiverWarningLevel: payload.readUInt16BE(10),
        minimumMessageInterval: payload.readUInt16BE(12),
        unknown: payload.readUInt16BE(14),
      }
      console.log("ICBM set channel", this.channel);
     return;
    }

    if (message.payload.subtype === 0x04) {
      const payload = Buffer.alloc(16, 0x00);
      payload.writeInt16BE(this.channel.channel, 0);
      payload.writeInt32BE(this.channel.messageFlags, 2);
      payload.writeInt16BE(this.channel.maxMessageSnacSize, 6);
      payload.writeInt16BE(this.channel.maxSenderWarningLevel, 8);
      payload.writeInt16BE(this.channel.maxReceiverWarningLevel, 10);
      payload.writeInt16BE(this.channel.minimumMessageInterval, 12);
      payload.writeInt16BE(this.channel.unknown, 14);

      // For some reason this response crashes the client?
      // It's identical to the channel set request the client
      // sends earlier. Also the 3.x client sends a channel set request
      // so early
      const resp = new FLAP(0x02, this.nextReqID,
        new SNAC(0x04, 0x05,  payload));
      this.send(resp);
      return;
    }
  }
}
