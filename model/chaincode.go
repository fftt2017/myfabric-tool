package model

import (
	"github.com/hyperledger/fabric/protos/peer"
)

type Chaincode struct {
	Name string
	Version string
}

func NewChaincode(cci *peer.ChaincodeInfo) *Chaincode {
	cc := new (Chaincode)
	cc.Name = cci.Name
	cc.Version = cci.Version
	return cc
}
