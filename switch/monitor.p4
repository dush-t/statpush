#include <core.p4>
#include <v1model.p4>

#include "header.p4"
#include "parser.p4"

/*
    CONSTANTS
    Change these to change how sensitive the switch is to
    detecting network congestion

    ECN_THRESHOLD -        The threshold length of the queue, after which we start looking for congestion   
    CON_THRESHOLD_TIME -   Time after which a queue length exceeding is classified as congestion. (In microseconds)
    DECON_THRESHOLD_TIME - Time after which an absence of queue length exceeded is classified as decongestion. (In microseconds) 2000000
*/
const bit<19> ECN_THRESHOLD = 1;
const bit<48> CON_THRESHOLD_TIME = 20000;
const bit<48> DECON_THRESHOLD_TIME = 20000;


control egress(inout headers hdr, inout metadata meta, inout standard_metadata_t standard_metadata) {

    action rewrite_mac(bit<48> smac) {
        hdr.ethernet.srcAddr = smac;
    }

    action _drop() {
        mark_to_drop(standard_metadata);
    }

    action send_to_controller() {
        
        clone(CloneType.E2E, 100);
    }

    table send_frame {
        actions = {
            rewrite_mac;
            _drop;
            NoAction;
        }
        key = {
            standard_metadata.egress_port: exact;
        }
        size = 256;
        default_action = NoAction();
    }

    apply {
        if (hdr.ipv4.isValid()) {
          send_frame.apply();
        }
    }
}

control ingress(inout headers hdr, inout metadata meta, inout standard_metadata_t standard_metadata) {
    
    register<bit<1>>(2) state;
    register<bit<48>>(2) timestamps;
    
    action _drop() {
        mark_to_drop(standard_metadata);
    }

    action set_nhop(bit<32> nhop_ipv4, bit<9> port) {
        meta.ingress_metadata.nhop_ipv4 = nhop_ipv4;
        standard_metadata.egress_spec = port;
        hdr.ipv4.ttl = hdr.ipv4.ttl + 8w255;
    }

    action set_dmac(bit<48> dmac) {
        hdr.ethernet.dstAddr = dmac;
    }

    action notify_controller() {
        meta.con_notif.queue_len = (bit<32>) standard_metadata.enq_qdepth;
        digest<con_notif_meta_t>(1, meta.con_notif);
    }

    table ipv4_lpm {
        actions = {
            _drop;
            set_nhop;
            NoAction;
        }
        key = {
            hdr.ipv4.dstAddr: lpm;
        }
        size = 1024;
        default_action = NoAction();
    }

    table forward {
        actions = {
            set_dmac;
            _drop;
            NoAction;
        }
        key = {
            meta.ingress_metadata.nhop_ipv4: exact;
        }
        size = 512;
        default_action = NoAction();
    }

    apply {

         /*  
            If the packet is not a cloned instance, take it into account while
            watching the queue length for congestion
        */

        bit<1> con;
        bit<1> qle;
        state.read(con, 0);
        state.read(qle, 1);
        
        if (con == 0 && qle == 0) {
            if (standard_metadata.enq_qdepth >= ECN_THRESHOLD) {
                state.write(1, 1);
                timestamps.write(0, standard_metadata.ingress_global_timestamp);
            }
        }
        else if (con == 0 && qle == 1) {
            if (standard_metadata.enq_qdepth >= ECN_THRESHOLD) {
                bit<48> init_timestamp;
                timestamps.read(init_timestamp, 0);
                if (standard_metadata.ingress_global_timestamp - init_timestamp > CON_THRESHOLD_TIME) {
                    state.write(0, 1);
                    // Send CON_START packet to controller here.
                    meta.con_notif.con_state = 1;
                    notify_controller();
                }
            }
            else {
                state.write(1, 0);
            }
        }
        else if (con == 1 && qle == 1) {
            // meta.con_notif.con_state = 1;
            // notify_controller();
            if (standard_metadata.enq_qdepth < ECN_THRESHOLD) {
                state.write(1, 0);
                timestamps.write(1, standard_metadata.ingress_global_timestamp);
            }
        }
        else if (con == 1 && qle == 0) {
            if (standard_metadata.enq_qdepth < ECN_THRESHOLD) {
                bit<48> end_timestamp;
                timestamps.read(end_timestamp, 1);
                if (standard_metadata.ingress_global_timestamp - end_timestamp > DECON_THRESHOLD_TIME) {
                    state.write(0, 0);
                    // Send CON_END packet to controller here
                    meta.con_notif.con_state = 0;
                    notify_controller();
                }
            }
        }

        if (hdr.ipv4.isValid()) {
          ipv4_lpm.apply();
          forward.apply();
        }
    }
}

V1Switch(ParserImpl(), verifyChecksum(), ingress(), egress(), computeChecksum(), DeparserImpl()) main;
