import net from 'net';
import Communicator from './communicator';

const server = net.createServer((socket) => {
  console.log('client connected...');
  socket.setTimeout(30000);

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

  new Communicator(socket);
});

server.on('error', (err) => {
  throw err;
});

server.listen(5190, () => {
  console.log('OSCAR ready on :5190');
});
