import BaseService from './base';
import Communicator from '../communicator';

export default class UsageStats extends BaseService {
  constructor(communicator : Communicator) {
    super({service: 0x0b, version: 0x01}, [], communicator)
  }
}
