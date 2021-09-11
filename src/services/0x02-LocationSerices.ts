import BaseService from './base';
import Communicator from '../communicator';

export default class LocationServices extends BaseService {
  constructor(communicator : Communicator) {
    super({family: 0x02, version: 0x01}, communicator)
  }
}
