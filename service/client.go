package service

import "myfabric-tool/chain/client"

func SetChainClient(ordererOrg, ordererNode, peerOrg, peerNode, userOrg, user string) error {
	return client.RefreshDefaultClient(ordererOrg, ordererNode, peerOrg, peerNode, userOrg, user)
}
