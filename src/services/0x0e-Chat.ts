import BaseService from './base';
import Communicator from '../communicator';

export default class Chat extends BaseService {
  constructor(communicator : Communicator) {
    super({service: 0x0e, version: 0x01}, communicator)
  }
}
