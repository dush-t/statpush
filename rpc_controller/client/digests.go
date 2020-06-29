package client

import (
	p4_v1 "github.com/p4lang/p4runtime/go/p4/v1"
)

// ConfigureDigest will take the required arguments and generate a DigestEntry message
// to configure how the switch treats the digest messages of a particular ID. Refer to
// the P4Runtime spec to see what each of these arguments/parameters do.
func (c *Client) ConfigureDigest(digestID uint32, maxListSize int32, maxTimeoutNs, ackTimeoutNs int64) error {
	digestConfig := &p4_v1.Update{
		Type: p4_v1.Update_INSERT,
		Entity: &p4_v1.Entity{
			Entity: &p4_v1.Entity_DigestEntry{
				// from build.bmv2/bmv2-digest.p4info.txt
				// digests { preamble { id: <DigestId> name "mac_learn_digest_t" } }
				DigestEntry: &p4_v1.DigestEntry{
					DigestId: digestID, //TODO: update based on p4info
					Config: &p4_v1.DigestEntry_Config{
						MaxTimeoutNs: maxTimeoutNs,
						MaxListSize:  maxListSize,
						AckTimeoutNs: ackTimeoutNs,
					},
				},
			},
		},
	}
	return c.WriteUpdate(digestConfig)
}
