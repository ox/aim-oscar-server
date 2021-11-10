import BaseService from './base';
import Communicator from '../communicator';

export default class Popups extends BaseService {
  constructor(communicator : Communicator) {
    super({service: 0x08, version: 0x01}, [], communicator)
  }
}
