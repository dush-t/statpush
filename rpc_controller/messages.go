package main

import (
	"github.com/dush-t/statpush/rpc_controller/client"
	log "github.com/sirupsen/logrus"
	"google.golang.org/genproto/googleapis/rpc/code"

	p4_v1 "github.com/p4lang/p4runtime/go/p4/v1"
)

// ListenToStreamMessages listens to messages received on the
// PullStreamMessages channel
func ListenToStreamMessages(c *client.Client) {
	go func() {
		for {
			request := &p4_v1.StreamMessageRequest{}
			c.PushStreamMessages <- request
			in := <-c.PullStreamMessages

			update := in.GetUpdate()

			switch update.(type) {
			case *p4_v1.StreamMessageResponse_Arbitration:
				HandleArbitration(in.GetArbitration())
			case *p4_v1.StreamMessageResponse_Digest:
				HandleDigest(in.GetDigest())
			default:
				log.Println("Message has unknown type")
			}
		}
	}()
}

// HandleDigest will handle the digest message
func HandleDigest(digest *p4_v1.DigestList) {
	messages := digest.GetData()

	var queueLen, timestamp []byte
	// var queueLen uint32
	// var timestamp uint64
	var conState uint

	for _, members := range messages {
		s := members.GetStruct()
		if s != nil {
			m := s.GetMembers()
			if bit := m[0].GetBitstring(); bit != nil {
				// queueLen = util.BinaryToUint32(bit)
				queueLen = bit
			}
			if bit := m[1].GetBitstring(); bit != nil {
				conState = uint(bit[0])
			}
			if bit := m[2].GetBitstring(); bit != nil {
				timestamp = bit
				// timestamp = util.Binary48ToInt64(bit)
			}
		}
	}
	if conState == 1 {
		log.Println("Congestion started")
	} else if conState == 0 {
		log.Println("Congestion ended")
	}
	log.Println("queueLen:", queueLen)
	log.Println("timeStamp:", timestamp)
	log.Println("###---------------------###")
}

// HandleArbitration will read the arbitration message and log
// whether or not this controller is master
func HandleArbitration(message *p4_v1.MasterArbitrationUpdate) {
	if message.Status.Code != int32(code.Code_OK) {
		log.Println("We are not master")
		// more handler code here
	} else {
		log.Println("We are master")
		// or here
	}
}
