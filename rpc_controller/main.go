package main

import (
	"context"
	"flag"
	"log"

	"google.golang.org/grpc"

	p4_v1 "github.com/p4lang/p4runtime/go/p4/v1"

	"github.com/dush-t/statpush/rpc_controller/client"
	"github.com/dush-t/statpush/rpc_controller/signals"
)

const (
	defaultAddr     = "127.0.0.1:50051"
	defaultDeviceID = 0
)

func main() {
	var binPath string
	flag.StringVar(&binPath, "bin", "", "Path to P4 bin")
	var p4InfoPath string
	flag.StringVar(&p4InfoPath, "p4info", "", "Path to p4info")

	flag.Parse()

	if binPath == "" || p4InfoPath == "" {
		log.Fatal("Missing .bin or P4Info")
	}

	conn, err := grpc.Dial(defaultAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal("Cannot connect to server", err)
	}

	defer conn.Close()

	c := p4_v1.NewP4RuntimeClient(conn)
	resp, err := c.Capabilities(context.Background(), &p4_v1.CapabilitiesRequest{})
	if err != nil {
		log.Fatal("Error in Capabilities RPC", err)
	}
	log.Println("P4Runtime server version is", resp.P4RuntimeApiVersion)

	stopCh := signals.RegisterSignalHandlers()

	electionId := p4_v1.Uint128{High: 0, Low: 1}

	p4RtC := client.NewClient(c, defaultDeviceID, electionId)
	startedCh := make(chan bool)
	go p4RtC.Run(stopCh, startedCh, binPath, p4InfoPath)

	// Wait for the client to finish starting up before performing
	// any read or write operations
	<-startedCh

	SetupSwitch(p4RtC)
	ListenToStreamMessages(p4RtC)

	log.Println("Press Ctrl-C to quit")
	<-stopCh
	log.Println("Stopping client")
}
