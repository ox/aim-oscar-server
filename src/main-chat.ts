import net from 'net';
import Communicator from './communicator';

import GenericServiceControls from "./services/0x01-GenericServiceControls";
import LocationServices from "./services/0x02-LocationSerices";
import BuddyListManagement from "./services/0x03-BuddyListManagement";
import ICBM from "./services/0x04-ICBM";
import Invitation from "./services/0x06-Invitation";
import Administration from "./services/0x07-Administration";
import Popups from "./services/0x08-Popups";
import PrivacyManagement from "./services/0x09-PrivacyManagement";
import UserLookup from "./services/0x0a-UserLookup";
import UsageStats from "./services/0x0b-UsageStats";
import ChatNavigation from "./services/0x0d-ChatNavigation";
import Chat from "./services/0x0e-Chat";;
import DirectorySearch from "./services/0x0f-DirectorySearch";
import ServerStoredBuddyIcons from "./services/0x10-ServerStoredBuddyIcons";
import SSI from "./services/0x13-SSI";

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
    new GenericServiceControls(comm),
    new LocationServices(comm),
    new BuddyListManagement(comm),
    new ICBM(comm),
    new Invitation(comm),
    new Administration(comm),
    new Popups(comm),
    new PrivacyManagement(comm),
    new UserLookup(comm),
    new UsageStats(comm),
    new ChatNavigation(comm),
    new Chat(comm),
    new DirectorySearch(comm),
    new ServerStoredBuddyIcons(comm),
    new SSI(comm),
  ];
  comm.registerServices(services);
  comm.startListening();
});

server.on('error', (err) => {
  throw err;
});

server.listen(5191, () => {
  console.log('CHAT ready on :5191');
});
