import BaseService from './base';
import Communicator from '../communicator';

export default class ChatNavigation extends BaseService {
  constructor(communicator : Communicator) {
    super({family: 0x0d, version: 0x02}, communicator)
  }
}
