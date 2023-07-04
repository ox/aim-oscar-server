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
- [x] Web Signup (https://runningman.network/register)
- [ ] Federation?

## Getting Started

Clone this repository and make sure you have [Go](https://go.dev/) installed in your terminal's path. Copy `env/example.config.yml` to `env/config.yml` and configure the service settings.

### OSCAR Settings

The server has two addresses that need to be set:

- `addr`: The host:port that the server listens on to provide Basic OSCAR Service

The `addr` needs to be an IP that the client can reach directly, not `0.0.0.0`. If you're running the client in a virtual environment then `addr` should be set to the local IP of the machine. On macOS you can find this by running:

```
osascript -e "IPv4 address of (system info)"
```

### Running

If this is the first time running this service you should do a DB migration to set up all of the tables and create a default user.

```
$ go run cmd/migrate/main.go --config <path to config> init
$ go run cmd/migrate/main.go --config <path to config> up
```

After you have set up your config you can run the server:

```
$ ./run.sh
```

If you set up your config somewhere else then set the `CONFIG_FILE` environment variable to the full path of the config file like so:

```
$ CONFIG_PATH=/Users/admin/config.yml ./run.sh
```

### Development

If you want to develop the aim-oscar-server, there is a `nodemon`-powered script in `./dev.sh` which will watch for changes and reload the aim-oscar-server automatically. The AIM clients are pretty good at not failing immediately when the server is unavailable so you can develop rapidly.

## User Administration

There is a user administration tool in `cmd/user` that lets you add and verify users on your server.

To add and verify a user:

```
$ go run cmd/user/main.go --config <path to config> add <screen_name> <password> <email>
```

To verify a user that has registered but not confirmed their email:

```
$ go run cmd/user/main.go --config <path to config> verify <screen_name>
```

### Terms

_from [iserverd](https://ox.github.io/iserverd-oscar-mirror/)_

- `BOS`: Basic OSCAR Service. This term refers to the services that form the core of the Instant Messenger service. These services include Login/Logoff, Locate, Instant Message, Roster management, Info management and Buddy List
- `FLAP` is a low-level communications protocol that facilitates the development of higher-level, record-oriented, communications layers. It is used on the TCP connection between all clients and servers.
- `SNAC`: A SNAC is the basic communication unit that is exchanged between clients and servers. The SNAC communication layers sits on top of the FLAP layer.
- `TLV`: Type Length Value. A tuple allowing typed opaque information to be passed through the protocol. Typically TLV's are intended for interpretation at the core layer. Being typed, new elements can be added w/o modifying the lower layers.
- `ICBM`: Inter Client Basic Message. ICBM is a channelized client-to-client mechanism. Currently the most user visible channel is used for Instant Messages.
