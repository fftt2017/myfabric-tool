package client

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric/common/configtx"
	localsigner "github.com/hyperledger/fabric/common/localmsp"
	"github.com/hyperledger/fabric/common/tools/protolator"
	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/scc/cscc"
	"github.com/hyperledger/fabric/core/scc/qscc"
	peercom "github.com/hyperledger/fabric/peer/common"
	protocom "github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/pkg/errors"
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
		return nil, fmt.Errorf("received bad response, status %d: %s", proposalResp.Response.Status, proposalResp.Response.Message)
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

func CreateForChannel(chainClient *ChainClient, channelTxFile, outputFile string, timeout time.Duration) error {
	cftx, err := ioutil.ReadFile(channelTxFile)
	if err != nil {
		return err
	}

	chCrtEnv, err := utils.UnmarshalEnvelope(cftx)
	if err != nil {
		return err
	}

	//sanityCheckAndSignConfigTx
	payload, err := utils.ExtractPayload(chCrtEnv)
	if err != nil {
		return err
	}

	if payload.Header == nil || payload.Header.ChannelHeader == nil {
		return err
	}

	ch, err := utils.UnmarshalChannelHeader(payload.Header.ChannelHeader)
	if err != nil {
		return err
	}

	if ch.Type != int32(protocom.HeaderType_CONFIG_UPDATE) {
		return errors.Errorf("bad type: %d", ch.Type)
	}

	if ch.ChannelId == "" {
		return errors.New("empty channel id")
	}

	configUpdateEnv, err := configtx.UnmarshalConfigUpdateEnvelope(payload.Data)
	if err != nil {
		return err
	}

	signer := localsigner.NewSigner()
	sigHeader, err := signer.NewSignatureHeader()
	if err != nil {
		return err
	}

	configSig := &protocom.ConfigSignature{
		SignatureHeader: utils.MarshalOrPanic(sigHeader),
	}

	configSig.Signature, err = signer.Sign(util.ConcatenateBytes(configSig.SignatureHeader, configUpdateEnv.ConfigUpdate))
	if err != nil {
		return err
	}

	configUpdateEnv.Signatures = append(configUpdateEnv.Signatures, configSig)

	chCrtEnv, err = utils.CreateSignedEnvelope(protocom.HeaderType_CONFIG_UPDATE, ch.ChannelId, signer, configUpdateEnv, 0, 0)
	if err != nil {
		return err
	}
	if err = chainClient.BroadcastClient.Send(chCrtEnv); err != nil {
		return err
	}

	//getGenesisBlock
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	deliverClient, err := peercom.NewDeliverClientForOrderer(ch.ChannelId)
	if err != nil {
		return err
	}

	var block *protocom.Block
	var stop = false
	for {
		select {
		case <-timer.C:
			deliverClient.Close()
			return errors.New("timeout waiting for channel creation")
		default:
			if block, err = deliverClient.GetSpecifiedBlock(0); err != nil {
				deliverClient.Close()
				deliverClient, err = peercom.NewDeliverClientForOrderer(ch.ChannelId)
				if err != nil {
					return err
				}
				time.Sleep(200 * time.Millisecond)
			} else {
				deliverClient.Close()
				stop = true
			}
		}
		if stop {
			break
		}
	}

	b, err := proto.Marshal(block)
	if err != nil {
		return err
	}

	file := ch.ChannelId + ".block"
	if outputFile != "" {
		file = outputFile
	}
	if err = ioutil.WriteFile(file, b, 0644); err != nil {
		return err
	}

	return nil
}

func JoinForChannel(chainClient *ChainClient, genesisBlockFile string) error {
	gb, err := ioutil.ReadFile(genesisBlockFile)
	if err != nil {
		return err
	}
	input := &pb.ChaincodeInput{Args: [][]byte{[]byte(cscc.JoinChain), gb}}
	spec := &pb.ChaincodeSpec{
		Type:        pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value["GOLANG"]),
		ChaincodeId: &pb.ChaincodeID{Name: "cscc"},
		Input:       input,
	}
	invocation := &pb.ChaincodeInvocationSpec{ChaincodeSpec: spec}

	creator, err := chainClient.Signer.Serialize()
	if err != nil {
		return err
	}

	var prop *pb.Proposal
	prop, _, err = utils.CreateProposalFromCIS(protocom.HeaderType_CONFIG, "", invocation, creator)
	if err != nil {
		return err
	}

	var signedProp *pb.SignedProposal
	signedProp, err = utils.GetSignedProposal(prop, chainClient.Signer)
	if err != nil {
		return err
	}

	var proposalResp *pb.ProposalResponse
	proposalResp, err = chainClient.EndorserClient.ProcessProposal(context.Background(), signedProp)
	if err != nil {
		return err
	}

	if proposalResp == nil {
		return errors.New("proposal resp is nil")
	}

	if proposalResp.Response.Status != 0 && proposalResp.Response.Status != 200 {
		return errors.Errorf("bad proposal response %d: %s", proposalResp.Response.Status, proposalResp.Response.Message)
	}

	return nil
}

func UpdateForChannel(chainClient *ChainClient, channelTxFile string) error {
	fileData, err := ioutil.ReadFile(channelTxFile)
	if err != nil {
		return err
	}

	ctxEnv, err := utils.UnmarshalEnvelope(fileData)
	if err != nil {
		return err
	}

	//sanityCheckAndSignConfigTx
	payload, err := utils.ExtractPayload(ctxEnv)
	if err != nil {
		return err
	}

	if payload.Header == nil || payload.Header.ChannelHeader == nil {
		return errors.New("bad header")
	}

	ch, err := utils.UnmarshalChannelHeader(payload.Header.ChannelHeader)
	if err != nil {
		return err
	}

	if ch.Type != int32(protocom.HeaderType_CONFIG_UPDATE) {
		return errors.Errorf("bad header: %d", ch.Type)
	}

	if ch.ChannelId == "" {
		return errors.New("empty channel id")
	}

	configUpdateEnv, err := configtx.UnmarshalConfigUpdateEnvelope(payload.Data)
	if err != nil {
		return err
	}

	signer := localsigner.NewSigner()
	sigHeader, err := signer.NewSignatureHeader()
	if err != nil {
		return err
	}

	configSig := &protocom.ConfigSignature{
		SignatureHeader: utils.MarshalOrPanic(sigHeader),
	}

	configSig.Signature, err = signer.Sign(util.ConcatenateBytes(configSig.SignatureHeader, configUpdateEnv.ConfigUpdate))
	if err != nil {
		return err
	}

	configUpdateEnv.Signatures = append(configUpdateEnv.Signatures, configSig)

	sCtxEnv, err := utils.CreateSignedEnvelope(protocom.HeaderType_CONFIG_UPDATE, ch.ChannelId, signer, configUpdateEnv, 0, 0)
	if err != nil {
		return err
	}
	if err = chainClient.BroadcastClient.Send(sCtxEnv); err != nil {
		return err
	}

	return nil
}
