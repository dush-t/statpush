from mininet.topo import Topo
from mininet.log import setLogLevel, info
from mininet.nodelib import LinuxBridge

import os

class Topology(Topo):

    """
    Creates a star topology with n hosts. This is heavily inspired from 
    p4app.
    """

    def __init__(self, sw_path, json_path, log_file,
                 thrift_port, pcap_dump, n, notifications_addr, **opts):
        Topo.__init__(self, **opts)

        setLogLevel('info')

        # Adding the switch
        switch = self.addSwitch('s1',
                                sw_path = sw_path,
                                json_path = json_path,
                                log_console = True,
                                log_file = log_file,
                                thrift_port = thrift_port,
                                enable_debugger = False,
                                pcap_dump = pcap_dump,
                                notifications_addr = notifications_addr)

        # Adding the hosts
        for h in xrange(n):
            host = self.addHost('h%d' % (h + 1),
                                ip = "10.0.%d.10/24" % h,
                                mac = "00:04:00:00:00:%02x" % h)
            info("Adding host %s\n" % str(host))
            self.addLink(host, switch)

        # Adding the controller
        # con_node = self.addHost("sw-cpu", cls=LinuxBridge, dpid='1000000000000000')
        # con_node = self.addHost('con',
        #                         ip = "10.1.1.10/24",
        #                         mac = "00:05:00:00:00:00")
        # self.addLink(switch, con_node, intfName1='s1-cpu-eth0', intfName2='s1-cpu-eth1')
        # self.addLink(con_node, switch)
    
        info('Topology set up\n')


