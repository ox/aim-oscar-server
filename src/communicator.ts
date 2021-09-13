import net from "net";
import { FLAP, SNAC, TLV, TLVType } from './structures';
import { logDataStream } from './util';
import { FLAGS_EMPTY } from './consts';

import GenericServiceControls from "./services/0x01-GenericServiceControls";
import LocationServices from "./services/0x02-LocationSerices";
import BuddyListManagement from "./services/0x03-BuddyListManagement";
import ICBM from "./services/0x04-ICBM";
import Invitation from "./services/0x06-Invitation";
import Administration from "./services/0x07-Administration";
import Popups from "./services/0x08-Popups";
import PrivacyManagement from "./services/0x09-PrivacyManagement";
import UserLookup from "./services/0x0a-UserLookup";
import UsageStats from "./services/0x0b-UsageStats";
import ChatNavigation from "./services/0x0d-ChatNavigation";
import Chat from "./services/0x0e-Chat";;
import DirectorySearch from "./services/0x0f-DirectorySearch";
import ServerStoredBuddyIcons from "./services/0x10-ServerStoredBuddyIcons";
import SSI from "./services/0x13-SSI";
import AuthorizationRegistrationService from "./services/0x17-AuthorizationRegistration";

import BaseService from "./services/base";

export interface User {
  uin: string,
  password: string,
  memberSince: Date,
}

export default class Communicator {

  private _sequenceNumber = 0;
  private messageBuffer = Buffer.alloc(0);
  public services : {[key: number]: BaseService} = {};
  public user? : User;

  constructor(public socket : net.Socket) {
    // Hold on to the socket
    this.socket = socket;

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

    this.registerServices();
    this.start();
  }

  start() {
    // Start negotiating a connection 
    const hello = new FLAP(0x01, 0, Buffer.from([0x00, 0x00, 0x00, 0x01]));
    this.send(hello);
  }

  registerServices() {
    const services = [
      new GenericServiceControls(this),
      new LocationServices(this),
      new BuddyListManagement(this),
      new ICBM(this),
      new Invitation(this),
      new Administration(this),
      new Popups(this),
      new PrivacyManagement(this),
      new UserLookup(this),
      new UsageStats(this),
      new ChatNavigation(this),
      new Chat(this),
      new DirectorySearch(this),
      new ServerStoredBuddyIcons(this),
      // new SSI(this),
      new AuthorizationRegistrationService(this),
    ];

    // Make a map of the service number to the service handler
    this.services = {};
    services.forEach((service) => {
      this.services[service.service] = service;
    });
  }

  _getNewSequenceNumber() {
    return ++this._sequenceNumber;
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
        console.log(tlv.toString());

        if (tlv.type === TLVType.GetServices) { // Requesting available services
          // this is just a dword list of subtype families
          const servicesOffered : Buffer[] = [];
          Object.values(this.services).forEach((subtype) => {
            servicesOffered.push(Buffer.from([0x00, subtype.service]));
          });
          const resp = new FLAP(2, this._getNewSequenceNumber(),
            new SNAC(0x01, 0x03, FLAGS_EMPTY, 0, Buffer.concat(servicesOffered)));
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
