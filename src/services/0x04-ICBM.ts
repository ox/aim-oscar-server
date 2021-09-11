import BaseService from './base';
import Communicator from '../communicator';

export default class ICBM extends BaseService {
  constructor(communicator : Communicator) {
    super({service: 0x04, version: 0x01}, communicator)
  }
}
