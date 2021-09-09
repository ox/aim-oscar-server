import Communicator from "../communicator";
import { FLAP } from "../structures";

interface ServiceFamilyVersion {
  family : number,
  version : number,
}

export default class BaseService {
  public family : number;
  public version : number;

  constructor({family, version} : ServiceFamilyVersion, public communicator : Communicator) {
    this.family = family;
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
    return;
  }
}
