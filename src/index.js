const net = require('net');
const Communicator = require('./communicator');

const communicators = {};

const server = net.createServer((socket) => {
  console.log('client connected...');
  socket.setTimeout(30000);

  socket.on('error', (e) => {
    console.error('socket encountered an error:', e);
    delete communicators[socket];
  });

  socket.on('timeout', () => {
    console.log('socket timeout');
    socket.end();
    delete communicators[socket];
  });

  socket.on('end', () => {
    console.log('client disconnected...');
  });

  const c = new Communicator(socket);
  communicators[socket] = c;
});

server.on('error', (err) => {
  throw err;
});

server.listen(5190, () => {
  console.log('OSCAR ready on :5190');
});
