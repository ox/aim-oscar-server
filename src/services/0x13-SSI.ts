import BaseService from './base';
import Communicator from '../communicator';

// SSI is Server Stored Information
export default class SSI extends BaseService {
  constructor(communicator : Communicator) {
    super({service: 0x13, version: 0x01}, communicator)
  }
}
