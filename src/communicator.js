const { FLAP, SNAC, TLV } = require('./structures');
const { logDataStream } = require('./util');
const { FLAGS_EMPTY } = require('./consts');

const AuthorizationRegistrationService = require("./services/authorization-registration");

class Communicator {
  constructor(socket) {
    // Hold on to the socket
    this.socket = socket;

    this.socket.on('data', (data) => {
      console.log('DATA-----------------------');
      console.log('RECV', logDataStream(Buffer.from(data, 'hex')));
      const flap = FLAP.fromBuffer(Buffer.from(data, 'hex'));
      console.log('RECV', flap.toString());
      this.handleMessage(flap);
    });

    this._sequenceNumber = 0;

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
      new AuthorizationRegistrationService(this),
    ];

    this.services = {};
    services.forEach((service) => {
      this.services[service.family] = service;
    });
  }

  _getNewSequenceNumber() {
    return ++this._sequenceNumber;
  }

  send(message) {
    console.log('SEND', message.toString());
    console.log('SEND', logDataStream(message.toBuffer()));
    console.log('-----------------------DATA');
    this.socket.write(message.toBuffer());
  }

  handleMessage(message) {
    switch (message.channel) {
      case 1:
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

        switch (tlv.type) {
          case 0x06: // Requesting available services
            // this is just a dword list of service families
            const servicesOffered = [];
            Object.values(this.services).forEach((service) => {
              servicesOffered.push(Buffer.from([0x00, service.family]));
            });
            const resp = new FLAP(2, this._getNewSequenceNumber(),
              new SNAC(0x01, 0x03, FLAGS_EMPTY, 0, [
                Buffer.concat(servicesOffered),
              ]));
            this.send(resp);
            return;
        }

        return;
      case 2:
        if (!message.payload) {
          console.error('No SNAC');
          return;
        }

        const familyService = this.services[message.payload.family];
        if (!familyService) {
          console.warn('no handler for family', message.payload.family);
          return;
        }

        familyService.handleMessage(message);
      default:
        console.warn('No handlers for channel', message.channel);
        return;
    }
  }
}

module.exports = Communicator;
