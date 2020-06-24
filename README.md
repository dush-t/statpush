# statpush
Programming network switches to notify controller about network congestion

## How to run
### Pre-requisites
To run this switch, you need to install the following dependencies - 
* P4 compiler
* BMv2 software switch
* Mininet

You can follow [these instructions](https://p4.org/p4/getting-started-with-p4.html) to install the compiler and the software switch. To install Mininet, you can follow the instructions on [this page](http://mininet.org/download/#option-2-native-installation-from-source). Alternatively, you can also install mininet using a package manager as demonstrated [here](http://mininet.org/download/#option-3-installation-from-packages). I installed it using the package manager and it seems to work fine. **Do not use the Mininet VM.**

**Note**: If you're on Ubuntu 20.04, there's a good chance that you'll encounter an error building the BMv2 switch. So I'd recommend using an Ubuntu 18.04 VM with at least 20GB storage (and not a Docker container). Alternatively you can also setup everything on an AWS EC2 instance, a DigitalOcean Droplet or a GCE VM instance. Just make sure you're setting up all this on Ubuntu 18.04 or below.

### Running the switch
Follow these steps - 
1. Clone this repository and navigate to it
```sh
git clone https://github.com/dush-t/statpush
cd statpush
```
2. Run run_network/run.py with Python2. You need to run this command with sudo because Mininet can only be run with sudo permissions.
```sh
sudo python2 run_network/run.py
```

If all goes well, you'll find yourself with the mininet CLI prompt. Try pinging a host to see if everything worked.

**Note**: The program might complain that the file */var/log/monitor.p4.log* does not exist. If it does, just create the file.

## Topology
This project uses a star topology with *n* hosts, where n can be changed from `run_network/conf.py`. If you wish to change the topology to something entirely different, you'll need to do it by editing the Topology class in `run_network/topology.py`.

You'll also need to make some additional changes to the python driver code if you want to run a non-star topology with multiple switches, but it's not quite reasonable to describe all those changes here. You'll need to read and understand all of the driver code (which isn't much, honestly) to be able to change the topology to a multi-switch one.

---

### Why not just use p4app?
I've written most of the driver code by hacking together different bits of the [p4app](github.com/p4lang/p4app) source code. I didn't use p4app directly because it does not allow setting a notification address to collect digests from the switch, at the moment. I needed this functionality earlier in this project, so I wrote the driver code myself for better control on the switch configuration.
