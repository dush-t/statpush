#!/usr/bin/env python
import argparse
import sys
import socket
import random
import struct
import time

from scapy.all import sendp, send, get_if_list, get_if_hwaddr
from scapy.all import Packet
from scapy.all import Ether, IP, UDP, TCP

def get_if():
    ifs=get_if_list()
    iface=None # "h1-eth0"
    for i in get_if_list():
        if "eth0" in i:
            iface=i
            break
    if not iface:
        print "Cannot find eth0 interface"
        exit(1)
    return iface

def main():

    if len(sys.argv) < 4:
        print 'pass 2 arguments: <ip> <port> <num_packets> <rate>'
        exit(1)




    addr = socket.gethostbyname(sys.argv[1])
    ip = sys.argv[1]
    port = int(sys.argv[2])
    num_packets = int(sys.argv[3])
    rate = float(sys.argv[4])
    iface = get_if()

    interval = 1/rate
    print interval

    print "sending on interface %s to %s" % (iface, str(addr))

    pkt =  Ether(src=get_if_hwaddr(iface),dst = "00:00:00:00:01:12")
    pkt = pkt /IP(dst=addr, tos=1)
    pkt.show2()

    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.connect((ip, port))

    for i in range(num_packets):
        sock.send("Hello from h1")
        print "Sent packet %s" % str(i)
        time.sleep(interval)

if __name__ == '__main__':
    main()
