import BaseService from './base';
import Communicator from '../communicator';

export default class UserLookup extends BaseService {
  constructor(communicator : Communicator) {
    super({service: 0x0a, version: 0x01}, communicator)
  }
}
