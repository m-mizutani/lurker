Lurker
===============

![The image under CC-BY-SA from Carbot Animation http://carbotanimations.wikia.com/wiki/Lurker](https://user-images.githubusercontent.com/605953/36069674-c55edbee-0f31-11e8-902e-c0793a80668b.png)

**Lurker** is a security network sensor like honeypot. The software should run on a host in internal network behind firewall, IPS or other network security protection products. Lurker need to run in a host that provides no service, then all other benign hosts in internal network should not access to the host. However an attacker who intruded to network will do port sweep scan and access to the host also where Lurker is running. Lurker detects a TCP SYN packet and reports it to security operators, and they can knows attacker's activities in their network. Additionally Lurker can reply a TCP SYN-ACK packet to gather a first TCP data packet from the TCP client for risk assesment and checking false positive.

By the way, The name of the software is inspired by a charactor of [Starcraft](http://us.battle.net/sc2/en/game/). The character burrows to underground and waits enemies silently.

Setup
--------------

```sh
$ mkdir -p $GOPATH/src/github.com/m-mizutani
$ cd $GOPATH/src/github.com/m-mizutani
$ git clone https://github.com/m-mizutani/lurker.git
$ cd lurker
$ dep ensure
$ go build
```

Run
---------------

Use case 1) Monitoring on eth0

```shell
$ sudo ./lurker -i eth0
```

Use case 2) Monitoring on eth0 and sending logs to fluentd server on 192.168.1.2 and port 24224

```shell
$ sudo ./lurker -i en0 -f 192.168.1.2:24224
```

Use case 3) Monitoring on eth0 and save logs to `lurker.log`

```shell
$ sudo ./lurker -i en0 -o lurker.log
```

Use case 4) Dry run with pcap file `test_data.pcap`

```shell
$ ./lurker -r test_data.pcap
```

License
---------------

- Source code: [BSD 2-Clause license](./LICENSE)
- [Image](https://user-images.githubusercontent.com/605953/36069674-c55edbee-0f31-11e8-902e-c0793a80668b.png): CC-BY-SA from Carbot Animation http://carbotanimations.wikia.com/wiki/Lurker

