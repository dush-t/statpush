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
const bit<48> CON_THRESHOLD_TIME = 1000000;
const bit<48> DECON_THRESHOLD_TIME = 1000000;


control egress(inout headers hdr, inout metadata meta, inout standard_metadata_t standard_metadata) {

    // These registers store information about the current switch
    // state.
    register<bit<1>>(2) state;
    register<bit<48>>(2) timestamps;

    action rewrite_mac(bit<48> smac) {
        hdr.ethernet.srcAddr = smac;
    }

    action _drop() {
        mark_to_drop(standard_metadata);
    }

    /*
        This method does not actually notify the controller, but recirculates
        the data packet with information about the congestion in the metadata.
        This metadata will be used in the ingress processing of the recirculated
        packet, where a digest will be generated and sent to the controller
        along with information about the congestion. 

        It's hacky. I know. But the thing is, there is no way to generate a digest
        in the egress pipeline, and no way to get queueing data in the ingress
        pipeline (since queueing happens after egress processing). Thus, I needed
        some way to somehow "call" an ingress action from the egress pipeline. Thus,
        the recirculation.

        Of course, I could just assume the existence of an extern that lets me
        specify some global switch state so I can pass data from egress to ingress,
        but that would make my code way less portable. And in the future if such an
        extern is available, all one needs to use it is to modify the action below.
    */
    action notify_controller() {
        meta.cloned_info.cloned = 1;
        meta.con_notif.queue_len = (bit<32>)standard_metadata.enq_qdepth;
        meta.con_notif.timestamp = standard_metadata.ingress_global_timestamp;
        meta.con_notif._padding = 5;
        recirculate(meta);
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
                    meta.con_notif.con_state = 1;
                    notify_controller();
                }
            }
            else {
                state.write(1, 0);
            }
        }
        else if (con == 1 && qle == 1) {
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
                    meta.con_notif.con_state = 0;
                    notify_controller();
                }
            }
        }
    }
}

control ingress(inout headers hdr, inout metadata meta, inout standard_metadata_t standard_metadata) {
    
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
            If the packet is recirculated, the below conditional will be
            evaluated to true and a digest will be sent to the controller.
        */
        if (meta.cloned_info.cloned == 1) {
            notify_controller();
            _drop();
        }

        if (hdr.ipv4.isValid()) {
          ipv4_lpm.apply();
          forward.apply();
        }
    }
}

V1Switch(ParserImpl(), verifyChecksum(), ingress(), egress(), computeChecksum(), DeparserImpl()) main;
