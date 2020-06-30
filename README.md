# statpush
Detecting network congestion in the data plane with P4 and Go

## How to run
### Pre-requisites
To run this switch, you need to install the following dependencies - 
* P4 compiler
* BMv2 software switch (`simple_switch_grpc`, specifically)
* P4Runtime
* Mininet
* GoLang

You can follow [these instructions](https://github.com/jafingerhut/p4-guide/blob/master/bin/README-install-troubleshooting.md) to install all the dependencies (except GoLang) with a few simple commands. I would recommend doing this on a fresh Ubuntu 18.04 VM, since the script on this link is regularly tested for Ubuntu 18.04.

**Note**: If you're on Ubuntu 20.04, there's a good chance that you'll encounter an error building the BMv2 switch. So I'd recommend using an Ubuntu 18.04 VM with at least 20GB storage (and not a Docker container). Alternatively you can also setup everything on an AWS EC2 instance, a DigitalOcean Droplet or a GCE VM instance. Just make sure you're setting up all this on Ubuntu 18.04 or below.

### Running the switch
Follow these steps. I'll be doing all these steps in `/home/dushyant`. 
1. Clone this repository and navigate to it
```sh
git clone https://github.com/dush-t/statpush
cd statpush
```
2. Run run_network/run.py with Python2. You need to run this command with sudo because Mininet can only be run with sudo permissions. This is going to be start a network with a star topology and an "empty" switch i.e. with no P4 program installed on it.
```sh
sudo python2 run_network/run.py
```
If all goes well, you'll find yourself with the mininet CLI prompt. Pinging hosts at this point will not work since the switch doesn't have a P4 binary installed on it (and doesn't have any table entries either).

3. Compile the P4 code to a P4Info object and the BMv2 JSON file. These will be used by our controller as API metadata to communicate with the switch through gRPC.
```sh
p4c --std p4_16 --p4runtime-files out/monitor.p4info.txt -o out switch/monitor.p4
```

4. Build and run the controller. The controller will install the P4 binary on the switch and populate the table entries through a gRPC connection.
```sh
cd rpc_controller
go build
sudo ./rpc_controller --bin /home/dushyant/statpush/out/monitor.json --p4info /home/dushyant/statpush/out/monitor.p4info.txt
```

5. Test the switch! If everything went well, running `h1 ping h2` will successfully send packets from `h1` to `h2`.

**Note**: The program might complain that the file */var/log/monitor.p4.log* does not exist. If it does, just create the file.

## Topology
This project uses a star topology with *n* hosts, where n can be changed from `run_network/conf.py`. If you wish to change the topology to something entirely different, you'll need to do it by editing the Topology class in `run_network/topology.py`.

You'll also need to make some additional changes to the python driver code if you want to run a non-star topology with multiple switches, but it's not quite reasonable to describe all those changes here. You'll need to read and understand all of the driver code (which isn't much, honestly) to be able to change the topology to a multi-switch one.

---

### Why not just use p4app?
I've written most of the driver code by hacking together different bits of the [p4app](github.com/p4lang/p4app) source code. I didn't use p4app directly because it does not allow setting a notification address to collect digests from the switch, at the moment. I needed this functionality earlier in this project, so I wrote the driver code myself for better control on the switch configuration.
