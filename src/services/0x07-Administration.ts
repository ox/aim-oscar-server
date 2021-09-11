import BaseService from './base';
import Communicator from '../communicator';

export default class Administration extends BaseService {
  constructor(communicator : Communicator) {
    super({service: 0x07, version: 0x01}, communicator)
  }
}
