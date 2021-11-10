import BaseService from './base';
import Communicator from '../communicator';

export default class Invitation extends BaseService {
  constructor(communicator : Communicator) {
    super({service: 0x06, version: 0x01}, [], communicator)
  }
}
