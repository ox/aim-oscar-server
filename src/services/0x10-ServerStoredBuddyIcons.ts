import BaseService from './base';
import Communicator from '../communicator';

export default class ServerStoredBuddyIcons extends BaseService {
  constructor(communicator : Communicator) {
    super({family: 0x10, version: 0x01}, communicator)
  }
}
