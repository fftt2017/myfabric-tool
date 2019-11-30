package model

import "github.com/hyperledger/fabric/protos/peer"

type Channel struct {
	Name string
}

func NewChannel(ci *peer.ChannelInfo) *Channel {
	c := new(Channel)
	c.Name = ci.GetChannelId()
	return c
}
