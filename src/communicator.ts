import net from "net";
import { FLAP, SNAC, TLV, TLVType } from './structures';
import { logDataStream } from './util';

import BaseService from "./services/base";

export interface User {
  uin: string,
  username: string,
  password: string,
  memberSince: Date,
}

export default class Communicator {

  private keepaliveInterval? : NodeJS.Timer;
  private _sequenceNumber = 0;
  private messageBuffer = Buffer.alloc(0);
  public services : {[key: number]: BaseService} = {};
  public user? : User;

  constructor(public socket : net.Socket) {}

  startListening() {
    this.socket.on('data', (data : Buffer) => {
      // we could get multiple FLAP messages, keep a running buffer of incoming
      // data and shift-off however many successful FLAPs we can make
      this.messageBuffer = Buffer.concat([this.messageBuffer, data]);

      while (this.messageBuffer.length > 0) {
        try {
          const flap = FLAP.fromBuffer(this.messageBuffer);
          console.log('DATA-----------------------');
          console.log('RAW\n' + logDataStream(flap.toBuffer()));
          console.log('RECV', flap.toString());
          this.messageBuffer = this.messageBuffer.slice(flap.length);
          this.handleMessage(flap);
          console.log('-----------------------DATA');
        } catch (e) {
          // Couldn't make a FLAP
          break;
        }
      }
    });

    this.keepaliveInterval = setInterval(() => {
      const keepaliveFlap = new FLAP(0x05, this.nextReqID, Buffer.from(""));
      this.socket.write(keepaliveFlap.toBuffer());
    }, 4 * 60 * 1000);

    this.socket.on('close', () => {
      if (this.keepaliveInterval) {
        clearInterval(this.keepaliveInterval);
      }
    });

    // Start negotiating a connection 
    const hello = new FLAP(0x01, 0, Buffer.from([0x00, 0x00, 0x00, 0x01]));
    this.send(hello);
  }

  registerServices(services : BaseService[] = []) {
    // Make a map of the service number to the service handler
    this.services = {};
    services.forEach((service) => {
      this.services[service.service] = service;
    });
  }

  get nextReqID() {
    return ++this._sequenceNumber & 0xFFFF;
  }

  send(message : FLAP) {
    console.log('SEND', message.toString());
    console.log('RAW\n' + logDataStream(message.toBuffer()));
    this.socket.write(message.toBuffer());
  }

  handleMessage(message : FLAP) {
    switch (message.channel) {
      case 1:
        // No SNACs on channel 1
        if (!(message.payload instanceof Buffer)) {
          return;
        }

        const protocol = message.payload.readUInt32BE();
        
        if (protocol !== 1) {
          console.log('Unsupported protocol:', protocol);
          this.socket.end();
          return;
        }

        if (message.payload.length <= 4) {
          return;
        }

        const tlv = TLV.fromBuffer(message.payload.slice(4));
        console.log('thing sent to channel 1:');
        console.log(tlv.toString());

        if (tlv.type === 0x06) {
          // client sent us a cookie
          const {cookie, user} = JSON.parse(tlv.payload.toString());
          console.log('cookie:', cookie);
          this.user = user;
        }

        if (tlv.type === TLVType.GetServices) { // Requesting available services
          // this is just a dword list of subtype families
          const servicesOffered : Buffer[] = [];
          Object.values(this.services).forEach((subtype) => {
            servicesOffered.push(Buffer.from([0x00, subtype.service]));
          });
          const resp = new FLAP(2, this.nextReqID,
            new SNAC(0x01, 0x03,  Buffer.concat(servicesOffered)));
          this.send(resp);
          return;
        }

        return;
      case 2:
        if (!(message.payload instanceof SNAC)) {
          console.error('Expected SNAC payload');
          return;
        }

        const familyService = this.services[message.payload.service];
        if (!familyService) {
          console.warn('no handler for service', message.payload.service);
          return;
        }

        familyService.handleMessage(message);
        return;
      default:
        console.warn('No handlers for channel', message.channel);
        return;
    }
  }
}
