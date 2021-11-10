import BaseService from './base';
import Communicator from '../communicator';

import { FLAGS_EMPTY } from '../consts';
import { FLAP, SNAC, TLV } from '../structures';
import { char, dword, word } from '../structures/bytes';

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
  /* Inter-Client Basic Message

  This system passes messages from/to clients through the server
  instead of directly between clients.
  */
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
    super({service: 0x04, version: 0x01}, [0x02, 0x04], communicator)
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
        unknown: 1000, //payload.readUInt16BE(14),
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

      const resp = new FLAP(0x02, this.nextReqID,
        new SNAC(0x04, 0x05,  payload));
      this.send(resp);
      return;
    }

    if (message.payload.subtype === 0x06) {
      // Client sent us a message to deliver to another client
      /*
        Channel 1 is used for what would commonly be called an "instant message" (plain text messages).
        Channel 2 is used for complex messages (rtf, utf8) and negotiating "rendezvous". These transactions end in something more complex happening, such as a chat invitation, or a file transfer.
        Channel 3 is used for chat messages (not in the same family as these channels).
      */

      if (!(message.payload.payload instanceof Buffer)) {
        throw new Error('this should be a buffer');
      }

      const msgId = message.payload.payload.readBigUInt64BE(0);
      const channel = message.payload.payload.readUInt16BE(8);
      const screenNameLength = message.payload.payload.readUInt8(10);
      const screenName = message.payload.payload.slice(11, 11 + screenNameLength).toString();

      console.log({
        msgId, channel, screenName,
      });

      if (channel === 1) {
        const tlvs = TLV.fromBufferBlob(message.payload.payload.slice(11 + screenNameLength));
        console.log(tlvs);

        // does the client want us to acknowledge that we got the message?
        const wantsAck = tlvs.find((tlv) => tlv.type === 3);

        // lets parse the message
        const messageTLV = tlvs.find((tlv) => tlv.type === 2);
        if (!messageTLV) {
          // TODO: send back error response
          throw new Error('need a message');
        }

        // Start parsing the message TLV payload
        // first is the array of capabilities
        const startOfMessageFragment = 2 + messageTLV.payload.readUInt16BE(2);
        const lengthOfMessageText = messageTLV.payload.readUInt16BE(startOfMessageFragment + 2);
        const messageText = messageTLV.payload.slice(startOfMessageFragment + 8, startOfMessageFragment + 8 + lengthOfMessageText).toString();
        console.log('The user said:', messageText);

        // The client usually wants a response that the server got the message. It checks that the message
        // back has the same message ID that was sent and the user it was sent to.
        if (wantsAck) {
          const sender = this.communicator.user?.username || "";
          const msgIdBuffer = Buffer.alloc(32);
          msgIdBuffer.writeBigUInt64BE(msgId);
          const ackPayload = Buffer.from([
            ...msgIdBuffer,
            ...word(0x02),
            ...char(sender.length),
            ...Buffer.from(sender),
          ]);
          const ackResp = new FLAP(2, this.nextReqID, new SNAC(0x04, 0x0c, ackPayload));
          this.send(ackResp);
        }
      }
    }
  }
}
