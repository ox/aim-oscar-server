import Communicator from "../communicator";
import { FLAP } from "../structures";

interface ServiceFamilyVersion {
  service : number,
  version : number,
}

export default class BaseService {
  public service : number;
  public version : number;

  constructor({service, version} : ServiceFamilyVersion, public communicator : Communicator) {
    this.service = service;
    this.version = version;
    this.communicator = communicator;
  }

  send(message : FLAP) {
    this.communicator.send(message);
  }

  _getNewSequenceNumber() {
    return this.communicator._getNewSequenceNumber();
  }

  handleMessage(message : FLAP) : void {
    throw new Error(''+
      `Unhandled message for service ${this.service.toString(16)} supporting version ${this.version.toString(16)}`);
  }
}
