import sys
import struct
import os
import time

from scapy.all import sniff, sendp, hexdump, get_if_list, get_if_hwaddr, bind_layers
from scapy.all import Packet, IPOption, Ether
from scapy.all import ShortField, IntField, LongField, BitField, FieldListField, FieldLenField
from scapy.all import IP, UDP, Raw, ls
from scapy.layers.inet import _IPOption_HDR

class ConNotifHeader(Packet):
    name = "ConNotifHeader"
    # fields_desc = [BitField("qid", 0, 8), BitField("queue_len", 0, 19), BitField("con_state", 0, 5)]
    fields_desc = [BitField("queue_len", 0, 19), BitField("con_state", 0, 1), BitField("_padding", 0, 4)]

bind_layers(ConNotifHeader, Ether)

def handle_packet(pkt):
    # print("Packet recieved")
    # print(pkt)
    p = ConNotifHeader(str(pkt))
    if p._padding == 5:
        if p.con_state == 1:
            print("Congestion started at queue. QueueLen=%s" % str(p.queue_len))
            p.show()
        elif p.con_state == 0:
            print("Congestion ended at queue. QueueLen=%s", str(p.queue_len))
            pass

    sys.stdout.flush()


def main():
    print("Flushing stdout")
    sys.stdout.flush()
    print("###############################")
    iface = 's1-eth3'
    print("Listening on %s" % iface)
    
    sniff(iface = iface, prn = lambda x: handle_packet(x))


if __name__ == '__main__':
    main()
