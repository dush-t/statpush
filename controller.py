import sys
import struct
import os

from scapy.all import sniff, bind_layers
from scapy.all import Packet, IPOption, Ether
from scapy.all import ShortField, IntField, LongField, BitField, FieldListField, FieldLenField
from scapy.all import IP, UDP, Raw, ls
from scapy.layers.inet import _IPOption_HDR

class IPHeader(Packet):
    name = 'ECNPacket'
    fields_desc = [
        BitField("version", 0, 4),
        BitField("ihl", 0, 4),
        BitField("diffserv", 0, 6),
        BitField("ecn", 0, 2),
        BitField("totalLen", 0, 16),
        BitField("identification", 0, 16),
        BitField("flags", 0, 3),
        BitField("fragOffset", 0, 13),
        BitField("ttl", 0, 8),
        BitField("protocol", 0, 8),
        BitField("hdrChecksum", 0, 16),
        BitField("srcAddr", 0, 32),
        BitField("dstAddr", 0, 32)
    ]

bind_layers(IPHeader, Ether)

def handle_pkt(pkt):
    print("Controller recieved a packet")

    p = IPHeader(str(pkt))
    if p.reason == 200:
        p.show()

    sys.stdout.flush()

def main():
    if len(sys.argv) < 2:
        iface = "s1-cpu-eth1"
    else:
        iface = sys.argv[1]

    print("Sniffing on %s" % iface)
    sys.stdout.flush()
    sniff(iface = iface, prn = lambda x: handle_pkt(x))

if __name__ == '__main__':
    main()