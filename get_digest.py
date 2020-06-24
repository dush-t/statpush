import nnpy
import struct
import time

class DigestController():

    def __init__(self, sw_id, con_handler, decon_handler):
        self.sw_id = sw_id
        self.con_handler = con_handler
        self.decon_handler = decon_handler

    def handle_msg_digest(self, msg):
        topic, device_id, ctx_id, list_id, buffer_id, num = struct.unpack("<iQiiQi", msg[:32])

        offset = 6
        msg = msg[32:]
        for sub_message in range(num):
            random_num, queue_len, con_state = struct.unpack("!BIB", msg[0:offset])
            info = {
                "sw_id": self.sw_id,
                "queue_len": queue_len,
                "timestamp": time.time()
            }
            
            if con_state == 1: self.con_handler(info)
            elif con_state == 0: self.decon_handler(info)
        
    def listen():
        sub = nnpy.Socket(nnpy.AF_SP, nnpy.SUB)
        sub.connect()
