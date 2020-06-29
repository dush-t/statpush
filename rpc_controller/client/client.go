package client

import (
	"context"

	p4_config_v1 "github.com/p4lang/p4runtime/go/p4/config/v1"
	p4_v1 "github.com/p4lang/p4runtime/go/p4/v1"
	log "github.com/sirupsen/logrus"
	"google.golang.org/genproto/googleapis/rpc/code"
)

// Client struct contains the data to represent a
// P4Runtime client
type Client struct {
	p4_v1.P4RuntimeClient
	deviceID           uint64
	electionID         p4_v1.Uint128
	p4Info             *p4_config_v1.P4Info
	PushStreamMessages chan *p4_v1.StreamMessageRequest
	PullStreamMessages chan *p4_v1.StreamMessageResponse
}

// NewClient will create a new P4Runtime client
func NewClient(p4RuntimeClient p4_v1.P4RuntimeClient, deviceID uint64, electionID p4_v1.Uint128) *Client {
	push := make(chan *p4_v1.StreamMessageRequest)
	pull := make(chan *p4_v1.StreamMessageResponse)

	return &Client{
		P4RuntimeClient:    p4RuntimeClient,
		deviceID:           deviceID,
		electionID:         electionID,
		PushStreamMessages: push,
		PullStreamMessages: pull,
	}
}

// Run will start the client's stream channel and will
// send an arbitration request to it
func (c *Client) Run(stopCh <-chan struct{}, startedCh chan bool, binPath, p4InfoPath string) error {
	log.Println("Starting the Stream Channel")

	// Start the Stream channel
	go c.StartStreamChannel()

	// Start the arbitration process
	request := &p4_v1.StreamMessageRequest{
		Update: &p4_v1.StreamMessageRequest_Arbitration{&p4_v1.MasterArbitrationUpdate{
			DeviceId:   c.deviceID,
			ElectionId: &c.electionID,
		}},
	}

	c.PushStreamMessages <- request

	arbitrationResponse := <-c.PullStreamMessages
	update := arbitrationResponse.GetUpdate()
	resp, ok := update.(*p4_v1.StreamMessageResponse_Arbitration)
	if !ok {
		log.Fatal("First stream message was not arbitration message. Wait, what?")
	}

	if resp.Arbitration.Status.Code != int32(code.Code_OK) {
		log.Println("We are not master. Wait, what?")
	} else {
		log.Println("We are master")
	}

	log.Println("Setting forwarding pipe")
	if err := c.SetFwdPipe(binPath, p4InfoPath); err != nil {
		log.Fatal("Error setting forwarding pipe", err)
	}

	// Tell the rest of the application that the client has
	// started i.e. arbitration is complete and the forwarding
	// pipe has been set
	startedCh <- true

	// Stop the client when a message is received on stopCh
	<-stopCh
	log.Println("Received stop signal")

	return nil
}

// WriteUpdate is used to update an entity on the
// switch. Refer to the P4Runtime spec to know more.
func (c *Client) WriteUpdate(update *p4_v1.Update) error {
	req := &p4_v1.WriteRequest{
		DeviceId:   c.deviceID,
		ElectionId: &c.electionID,
		Updates:    []*p4_v1.Update{update},
	}

	_, err := c.Write(context.Background(), req)
	return err
}
