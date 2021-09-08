class BaseService {
  constructor({family, version}, communicator) {
    this.family = family;
    this.version = version;
    this.communicator = communicator;
  }

  send(message) {
    this.communicator.send(message);
  }

  _getNewSequenceNumber() {
    return this.communicator._getNewSequenceNumber();
  }

  handleMessage(message) {
    return null;
  }
}

module.exports = BaseService;
