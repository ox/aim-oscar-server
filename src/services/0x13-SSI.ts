import BaseService from './base';
import Communicator from '../communicator';

// SSI is Server Stored Information
export default class SSI extends BaseService {
  constructor(communicator : Communicator) {
    super({family: 0x10, version: 0x01}, communicator)
  }
}
