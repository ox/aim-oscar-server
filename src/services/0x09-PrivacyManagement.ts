import BaseService from './base';
import Communicator from '../communicator';

export default class PrivacyManagement extends BaseService {
  constructor(communicator : Communicator) {
    super({service: 0x09, version: 0x01}, communicator)
  }
}
