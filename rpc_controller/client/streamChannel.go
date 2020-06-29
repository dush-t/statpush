package client

import (
	"context"
	"io"

	log "github.com/sirupsen/logrus"
)

// StartStreamChannel starts two goroutines - one listens to the stream
// channel and sends any messages received to the PullStreamMessages
// channel, and the other listens for messages on the PushStreamMessages
// channel and sends them to the stream. So in short, to send any stream
// messages the developer can send them to the PushStreamMessages channel
// and to receive any stream messages the developer can listen to the
// PullStreamMessages channel
func (c *Client) StartStreamChannel() {
	stream, err := c.StreamChannel(context.Background())
	if err != nil {
		log.Fatal("Cannot establish stream channel:", err)
	}

	// Start a goroutine that gets messages from the stream and sends those
	// messages to the PullStreamMessages channel
	go func() {
		for {
			in, err := stream.Recv()
			log.Println("Received a message from stream channel")
			if err == io.EOF {
				log.Println("Error receiving message from stream:", err)
			}
			if err != nil {
				// log.Fatalf("Failed to receive a note : %v", err)
				log.Println("Error receiving message from stream:", err)
			}

			c.PullStreamMessages <- in
		}
	}()

	// Start a goroutine that listens on PushStreamMessages, and sends any message
	// received on that channel to the gRPC stream channel
	go func() {
		for {
			sendmess := <-c.PushStreamMessages
			log.Println("Sending message to stream channel")
			mess := sendmess
			if err := stream.Send(mess); err != nil {
				log.Println("Unable to send message to stream channel:", err)
			}
		}
	}()
}
