import net from "net";

const server = net.createServer((socket) => {
  socket.setTimeout(5 * 60 * 1000); // 5 minute timeout
  socket.on('timeout', () => {
    console.log('socket timeout');
    socket.end();
  });

  socket.on('end', () => {
    console.log('client disconnected...');
  });
});

server.on('error', (err) => {
  console.error(err);
});

server.listen(9999, () => {
  console.log('proxy ready');
})
