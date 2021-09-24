import net from 'net';
import Communicator from './communicator';

import AuthorizationRegistrationService from "./services/0x17-AuthorizationRegistration";

const server = net.createServer((socket) => {
  console.log('client connected...');
  socket.setTimeout(5 * 60 * 1000); // 5 minute timeout

  socket.on('error', (e) => {
    console.error('socket encountered an error:', e);
    socket.end();
  });

  socket.on('timeout', () => {
    console.log('socket timeout');
    socket.end();
  });

  socket.on('end', () => {
    console.log('client disconnected...');
  });

  const comm = new Communicator(socket);
  const services = [
    new AuthorizationRegistrationService(comm),
  ];
  comm.registerServices(services);
  comm.startListening();
});

server.on('error', (err) => {
  throw err;
});

server.listen(5190, () => {
  console.log('AUTH ready on :5190');
});
