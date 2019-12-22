package client

import (
	"bytes"
	"context"
	"fmt"
	"strconv"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric/common/tools/protolator"
	"github.com/hyperledger/fabric/core/scc/cscc"
	"github.com/hyperledger/fabric/core/scc/qscc"
	peercom "github.com/hyperledger/fabric/peer/common"
	protocom "github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
)

func ListChannels(chainClient *ChainClient) ([]*pb.ChannelInfo, error) {
	invocation := &pb.ChaincodeInvocationSpec{
		ChaincodeSpec: &pb.ChaincodeSpec{
			Type:        pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value["GOLANG"]),
			ChaincodeId: &pb.ChaincodeID{Name: "cscc"},
			Input:       &pb.ChaincodeInput{Args: [][]byte{[]byte(cscc.GetChannels)}},
		},
	}
	c, _ := chainClient.Signer.Serialize()

	prop, _, err := utils.CreateProposalFromCIS(protocom.HeaderType_ENDORSER_TRANSACTION, "", invocation, c)
	if err != nil {
		return nil, err
	}
	//fmt.Println(prop)
	signedProp, err := utils.GetSignedProposal(prop, chainClient.Signer)
	if err != nil {
		return nil, err
	}
	//fmt.Println(signedProp)
	proposalResp, err := chainClient.EndorserClient.ProcessProposal(context.Background(), signedProp)
	if err != nil {
		return nil, err
	}
	//fmt.Println(proposalResp)
	if proposalResp.Response == nil || proposalResp.Response.Status != 200 {
		return nil, fmt.Errorf("Received bad response, status %d: %s", proposalResp.Response.Status, proposalResp.Response.Message)
	}
	var channelQueryResponse pb.ChannelQueryResponse
	err = proto.Unmarshal(proposalResp.Response.Payload, &channelQueryResponse)
	if err != nil {
		return nil, err
	}
	return channelQueryResponse.Channels, nil
}

func GetChannelInfo(chainClient *ChainClient, channelID string) (*protocom.BlockchainInfo, error) {
	invocation := &pb.ChaincodeInvocationSpec{
		ChaincodeSpec: &pb.ChaincodeSpec{
			Type:        pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value["GOLANG"]),
			ChaincodeId: &pb.ChaincodeID{Name: "qscc"},
			Input:       &pb.ChaincodeInput{Args: [][]byte{[]byte(qscc.GetChainInfo), []byte(channelID)}},
		},
	}
	c, _ := chainClient.Signer.Serialize()

	prop, _, err := utils.CreateProposalFromCIS(protocom.HeaderType_ENDORSER_TRANSACTION, "", invocation, c)
	if err != nil {
		return nil, err
	}
	//fmt.Println(prop)
	signedProp, err := utils.GetSignedProposal(prop, chainClient.Signer)
	if err != nil {
		return nil, err
	}
	//fmt.Println(signedProp)
	proposalResp, err := chainClient.EndorserClient.ProcessProposal(context.Background(), signedProp)
	if err != nil {
		return nil, err
	}
	if proposalResp.Response == nil || proposalResp.Response.Status != 200 {
		return nil, fmt.Errorf("received bad response, status %d: %s \n", proposalResp.Response.Status, proposalResp.Response.Message)
	}
	blockChainInfo := &protocom.BlockchainInfo{}
	err = proto.Unmarshal(proposalResp.Response.Payload, blockChainInfo)
	if err != nil {
		return nil, err
	}
	return blockChainInfo, nil
}

func FetchBlock(chainClient *ChainClient, channelID, blockIndex string) (string, error) {
	deliverClient, err := peercom.NewDeliverClientForPeer(channelID)
	if err != nil {
		return "", err
	}
	var block *protocom.Block
	switch blockIndex {
	case "oldest":
		block, err = deliverClient.GetOldestBlock()
	case "newest":
		block, err = deliverClient.GetNewestBlock()
	case "config":
		iBlock, err := deliverClient.GetNewestBlock()
		if err != nil {
			return "", err
		}
		lc, err := utils.GetLastConfigIndexFromBlock(iBlock)
		if err != nil {
			return "", err
		}
		block, err = deliverClient.GetSpecifiedBlock(lc)
	default:
		num, err := strconv.Atoi(blockIndex)
		if err != nil {
			return "", err
		}
		block, err = deliverClient.GetSpecifiedBlock(uint64(num))
	}
	if err != nil {
		return "", err
	}
	bs := make([]byte, 0, 10*1024)
	buf := bytes.NewBuffer(bs)
	err = protolator.DeepMarshalJSON(buf, block)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
