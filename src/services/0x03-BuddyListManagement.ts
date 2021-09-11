import BaseService from './base';
import Communicator from '../communicator';

export default class BuddyListManagement extends BaseService {
  constructor(communicator : Communicator) {
    super({service: 0x03, version: 0x01}, communicator)
  }
}
