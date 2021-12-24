# AIM Oscar Server

Run your own AIM chat server, managing users and groups. Hook up a vintage client and get chatty.

## Goals

- [x] Have a vintage client authenticate with the server
- [x] Add buddies
- [x] See buddy online/away status
- [x] Chat with buddy
- [x] Set away status
- [ ] See away status
- [ ] Look up buddy
- [ ] Buddy icons
- [ ] Rate limiting + warn system
- [ ] Web Signup
- [ ] Federation with other servers

## Getting Started

Clone this repository, make sure you have [Go](https://go.dev/) installed in your terminal's path, then run:

```
$ go build && ./aim-oscar
```

### Configuration

Environment flags:

- OSCAR_HOST: hostname of the server
- OSCAR_PORT: port to bind to
- OSCAR_BOS_HOST: hostname of Basic OSCAR Service that provides core client features
- OSCAR_BOS_PORT: port of Basic OSCAR Service

Flags:

- `-host`: hostname of server
- `-port`: port to bind to
- `-boshost`: hostname of Basic OSCAR Service
- `-bosort`: port of Basic OSCAR Service
- `-h`: see help information about flags

### Terms

_mirrored from [iserverd](http://iserverd.khstu.ru/oscar/terms.html)_

- `BOS`: Basic OSCAR Service. This term refers to the services that form the core of the Instant Messenger service. These services include Login/Logoff, Locate, Instant Message, Roster management, Info management and Buddy List
- `FLAP` is a low-level communications protocol that facilitates the development of higher-level, record-oriented, communications layers. It is used on the TCP connection between all clients and servers.
- `SNAC`: A SNAC is the basic communication unit that is exchanged between clients and servers. The SNAC communication layers sits on top of the FLAP layer.
- `TLV`: Type Length Value. A tuple allowing typed opaque information to be passed through the protocol. Typically TLV's are intended for interpretation at the core layer. Being typed, new elements can be added w/o modifying the lower layers.
- `ICBM`: Inter Client Basic Message. ICBM is a channelized client-to-client mechanism. Currently the most user visible channel is used for Instant Messages.
