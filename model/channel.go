package model

import (
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/peer"
)

type Channel struct {
	Name string
}

type Blockchain struct {
	Height            uint64
	CurrentBlockHash  []byte
	PreviousBlockHash []byte
}

func NewChannel(ci *peer.ChannelInfo) *Channel {
	c := new(Channel)
	c.Name = ci.GetChannelId()
	return c
}

func NewBlockchain(bi *common.BlockchainInfo) *Blockchain {
	b := new(Blockchain)
	b.Height = bi.Height
	b.CurrentBlockHash = bi.CurrentBlockHash
	bi.PreviousBlockHash = bi.PreviousBlockHash
	return b
}
