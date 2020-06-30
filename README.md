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

### Running the network
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

3. Compile the P4 code to a P4Info object and the BMv2 JSON file (yes, there's no issue with doing this step _after_ running the network). These will be used by our controller as API metadata to communicate with the switch through gRPC.
```sh
p4c --std p4_16 --p4runtime-files out/monitor.p4info.txt -o out switch/monitor.p4
```
### Running the controller
The controller is implemented in GoLang. You first need to build it, and then run it. Running the controller will install the P4 binaries and table entries on the switch by sending it protocol buffers over gRPC. Ensure that the network is running before you start the controller.
```sh
cd rpc_controller
go build
sudo ./rpc_controller --bin /home/dushyant/statpush/out/monitor.json --p4info /home/dushyant/statpush/out/monitor.p4info.txt
```

### Test the switch!
If everything went well, running `h1 ping h2` will successfully send packets from `h1` to `h2`.

**Note**: The program might complain that the file */var/log/monitor.p4.log* does not exist. If it does, just create the file.

## Topology
This project uses a star topology with *n* hosts, where n can be changed from `run_network/conf.py`. If you wish to change the topology to something entirely different, you'll need to do it by editing the Topology class in `run_network/topology.py`.

You'll also need to make some additional changes to the python driver code if you want to run a non-star topology with multiple switches, but it's not quite reasonable to describe all those changes here. You'll need to read and understand all of the driver code (which isn't much, honestly) to be able to change the topology to a multi-switch one.

## Some issues

### Sending digest messages in the egress pipeline
In the `v1model` PSA architecture, packets are queued after ingress processing. On the other hand, digests messages can be sent to the control plane only during the ingress processing. Thus, simply sending digest messages based on queueing-related metadata is impossible. It's an architecture limitation and there is nothing that can be done to achieve it while using the `v1model` architecture.

However, in this project I have implemented a hacky (and limited) fix. I do all the processing in the egress pipeline, and if I need to send a digest message, I simply recirculate the packet with the appropriate metadata. In the ingress pipeline, I check if the packet is a recirculated one (using it's metadata), and if it is, I send that metadata in a digest message and then drop the packet. This is a rather hacky way to "call" an ingress action from the egress pipeline, but it works.

The issue with this is that sometimes it _does not_ work, and this is because of a [fundamental bug](https://github.com/p4lang/behavioral-model/blob/master/docs/simple_switch.md#restrictions-on-recirculate-resubmit-and-clone-operations) in how `p4c` handles preserving metadata in `v1model` during `recirculate` and `resubmit` operations. Not much I can do about this. 

### BMv2's poor performance
BMv2 _is not_ a production grade switch. It is for testing purposes only. So when I try to hit it with heavy traffic (think 1GBps), most of the packets are just lost and the logs show that the queue depth never exceeded the threshold. Thus, congestion is _not always_ detected. This failure in congestion detection, however, is not relevant to the scope of the project because in this case the switch itself does not know of congestion (while the project focusses on notifying the controller _once the switch finds out_). I have verified from the switch logs, however, that whenever the switch does detect congestion, it notifies the controll plane.

This should not be a problem in production grade switches which don't just lose packets.

---

### Why not just use p4app?
I've written most of the driver code by hacking together different bits of the [p4app](github.com/p4lang/p4app) source code. I didn't use p4app directly because it does not allow setting a notification address to collect digests from the switch, at the moment. I needed this functionality earlier in this project, so I wrote the driver code myself for better control on the switch configuration.
