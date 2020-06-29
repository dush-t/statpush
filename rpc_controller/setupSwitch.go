package main

import (
	"log"

	"github.com/dush-t/statpush/rpc_controller/client"
	"github.com/dush-t/statpush/rpc_controller/util"
)

func addSendFrameEntry(c *client.Client, port uint32, mac string) error {
	byteMac, err := util.MacToBinary(mac)
	bytePort, err := util.UInt32ToBinary(port, 2)
	if err != nil {
		return err
	}

	err1 := c.InsertTableEntry("egress.send_frame", "egress.rewrite_mac",
		[]client.MatchInterface{&client.ExactMatch{
			Value: bytePort,
		}},
		[]([]byte){byteMac},
	)

	if err1 != nil {
		return err1
	}

	return nil
}

func addForwardEntry(c *client.Client, ip string, mac string) error {
	byteIp, err := util.IpToBinary(ip)
	byteMac, err1 := util.MacToBinary(mac)
	if err != nil || err1 != nil {
		return err
	}

	err2 := c.InsertTableEntry("ingress.forward", "ingress.set_dmac",
		[]client.MatchInterface{&client.ExactMatch{
			Value: byteIp,
		}},
		[]([]byte){byteMac},
	)

	if err2 != nil {
		return err1
	}

	return nil
}

func addIpv4LpmEntry(c *client.Client, ip string, port uint32) error {
	byteIp, err := util.IpToBinary(ip)
	bytePort, err := util.UInt32ToBinary(port, 2)

	if err != nil {
		return err
	}

	err1 := c.InsertTableEntry("ingress.ipv4_lpm", "ingress.set_nhop",
		[]client.MatchInterface{&client.LpmMatch{
			Value: byteIp,
			PLen:  32,
		}},
		[]([]byte){byteIp, bytePort},
	)

	if err1 != nil {
		return err1
	}

	return nil
}

func SetupSwitch(c *client.Client) {

	err := addSendFrameEntry(c, 1, "00:aa:bb:00:00:00")
	err = addSendFrameEntry(c, 2, "00:aa:bb:00:00:01")

	err = addForwardEntry(c, "10.0.0.10", "00:04:00:00:00:00")
	err = addForwardEntry(c, "10.0.1.10", "00:04:00:00:00:01")

	err = addIpv4LpmEntry(c, "10.0.0.10", 1)
	err = addIpv4LpmEntry(c, "10.0.1.10", 2)

	if err != nil {
		log.Fatal(err)
	}

	e := c.ConfigureDigest(393173492, 1, 0, 0)
	if e != nil {
		log.Println(e)
	}
}
