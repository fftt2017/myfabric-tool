package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/hyperledger/fabric/core/common/ccprovider"
	peercom "github.com/hyperledger/fabric/peer/common"
	protocom "github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/pkg/errors"
)

const (
	ChaincodeLang = "golang"
)

func ListInstalledForChaincode(chainClient *ChainClient) ([]*pb.ChaincodeInfo, error) {
	creator, err := chainClient.Signer.Serialize()
	if err != nil {
		return nil, err
	}
	prop, _, err := utils.CreateGetInstalledChaincodesProposal(creator)
	if err != nil {
		return nil, err
	}
	signedProp, err := utils.GetSignedProposal(prop, chainClient.Signer)
	if err != nil {
		return nil, err
	}
	proposalResp, err := chainClient.EndorserClient.ProcessProposal(context.Background(), signedProp)
	if err != nil {
		return nil, err
	}
	if proposalResp.Response == nil || proposalResp.Response.Status != 200 {
		return nil, fmt.Errorf("received bad response, status %d: %s", proposalResp.Response.Status, proposalResp.Response.Message)
	}

	cqr := &pb.ChaincodeQueryResponse{}

	err = proto.Unmarshal(proposalResp.Response.Payload, cqr)
	if err != nil {
		return nil, err
	}
	return cqr.Chaincodes, nil
}

func ListInstantiatedForChaincode(chainClient *ChainClient, channelID string) ([]*pb.ChaincodeInfo, error) {
	creator, err := chainClient.Signer.Serialize()
	if err != nil {
		return nil, err
	}
	prop, _, err := utils.CreateGetChaincodesProposal(channelID, creator)
	if err != nil {
		return nil, err
	}
	signedProp, err := utils.GetSignedProposal(prop, chainClient.Signer)
	if err != nil {
		return nil, err
	}
	proposalResp, err := chainClient.EndorserClient.ProcessProposal(context.Background(), signedProp)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	if proposalResp.Response == nil || proposalResp.Response.Status != 200 {
		return nil, fmt.Errorf("received bad response, status %d: %s", proposalResp.Response.Status, proposalResp.Response.Message)
	}
	cqr := &pb.ChaincodeQueryResponse{}
	err = proto.Unmarshal(proposalResp.Response.Payload, cqr)
	if err != nil {
		return nil, err
	}
	return cqr.Chaincodes, nil
}

func QueryForChaincode(chainClient *ChainClient, channelID, chaincodeName, args string) (string, error) {
	spec := &pb.ChaincodeSpec{}
	input := &pb.ChaincodeInput{}
	if err := json.Unmarshal([]byte(args), &input); err != nil {
		return "", nil
	}
	chaincodeLang := strings.ToUpper(ChaincodeLang)
	spec = &pb.ChaincodeSpec{
		Type:        pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value[chaincodeLang]),
		ChaincodeId: &pb.ChaincodeID{Path: peercom.UndefinedParamValue, Name: chaincodeName, Version: peercom.UndefinedParamValue},
		Input:       input,
	}

	invocation := &pb.ChaincodeInvocationSpec{ChaincodeSpec: spec}
	creator, err := chainClient.Signer.Serialize()
	if err != nil {
		return "", nil
	}
	var tMap map[string][]byte
	txID := ""
	prop, txid, err := utils.CreateChaincodeProposalWithTxIDAndTransient(protocom.HeaderType_ENDORSER_TRANSACTION, channelID, invocation, creator, txID, tMap)
	if err != nil {
		return "", nil
	}
	_ = txid
	signedProp, err := utils.GetSignedProposal(prop, chainClient.Signer)
	if err != nil {
		return "", nil
	}
	proposalResp, err := chainClient.EndorserClient.ProcessProposal(context.Background(), signedProp)
	if err != nil {
		return "", nil
	}
	//fmt.Println(proposalResp)
	if proposalResp == nil {
		return "", fmt.Errorf("error during query: received nil proposal response")
	}
	if proposalResp.Endorsement == nil {
		return "", fmt.Errorf("endorsement failure during query. response: %v", proposalResp.Response)
	}
	return string(proposalResp.Response.Payload), nil
}

func InvokeForChainCode(chainClient *ChainClient, channelID, chaincodeName, args string) error {
	spec := &pb.ChaincodeSpec{}
	input := &pb.ChaincodeInput{}
	if err := json.Unmarshal([]byte(args), &input); err != nil {
		return err
	}
	chaincodeLang := strings.ToUpper(ChaincodeLang)
	spec = &pb.ChaincodeSpec{
		Type:        pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value[chaincodeLang]),
		ChaincodeId: &pb.ChaincodeID{Path: peercom.UndefinedParamValue, Name: chaincodeName, Version: peercom.UndefinedParamValue},
		Input:       input,
	}
	invocation := &pb.ChaincodeInvocationSpec{ChaincodeSpec: spec}
	creator, err := chainClient.Signer.Serialize()
	if err != nil {
		return err
	}
	var tMap map[string][]byte
	txID := ""
	prop, txid, err := utils.CreateChaincodeProposalWithTxIDAndTransient(protocom.HeaderType_ENDORSER_TRANSACTION, channelID, invocation, creator, txID, tMap)
	if err != nil {
		return err
	}
	fmt.Println("txid: ", txid)
	signedProp, err := utils.GetSignedProposal(prop, chainClient.Signer)
	if err != nil {
		return err
	}
	proposalResp, err := chainClient.EndorserClient.ProcessProposal(context.Background(), signedProp)
	if err != nil {
		return err
	}
	//var responses []*pb.ProposalResponse
	//responses = append(responses, proposalResp)
	env, err := utils.CreateSignedTx(prop, chainClient.Signer, proposalResp)
	if err != nil {
		return err
	}

	if err = chainClient.BroadcastClient.Send(env); err != nil {
		return err
	}

	return nil
}

func InstallForChaincode(chainClient *ChainClient, chaincodeName, chaincodeVersion, chaincodePath, codePackage string) error {
	if existed, _ := ccprovider.ChaincodePackageExists(chaincodeName, chaincodeVersion); existed {
		return errors.Errorf("chaincode %s:%s already exists", chaincodeName, chaincodeVersion)
	}
	spec := &pb.ChaincodeSpec{}
	input := &pb.ChaincodeInput{}
	chaincodeCtorJSON := "{}"
	if err := json.Unmarshal([]byte(chaincodeCtorJSON), &input); err != nil {
		return err
	}
	spec = &pb.ChaincodeSpec{
		Type:        pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value[strings.ToUpper(ChaincodeLang)]),
		ChaincodeId: &pb.ChaincodeID{Path: chaincodePath, Name: chaincodeName, Version: chaincodeVersion},
		Input:       input,
	}

	codePackageBytes, err := ioutil.ReadFile(codePackage)
	if err != nil {
		return err
	}
	chaincodeDeploymentSpec := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: spec, CodePackage: codePackageBytes}

	creator, err := chainClient.Signer.Serialize()
	if err != nil {
		return err
	}

	prop, _, err := utils.CreateInstallProposalFromCDS(chaincodeDeploymentSpec, creator)
	if err != nil {
		return err
	}

	var signedProp *pb.SignedProposal
	signedProp, err = utils.GetSignedProposal(prop, chainClient.Signer)
	if err != nil {
		return err
	}

	// install is currently only supported for one peer
	proposalResponse, err := chainClient.EndorserClient.ProcessProposal(context.Background(), signedProp)
	if err != nil {
		return err
	}

	if proposalResponse == nil {
		return errors.New("Error during install: received nil proposal response")
	}

	if proposalResponse.Response.Status != int32(protocom.Status_SUCCESS) {
		return errors.Errorf("Bad response: %d - %s", proposalResponse.Response.Status, proposalResponse.Response.Message)
	}

	return nil
}

func InstantiateForChaincode(chainClient *ChainClient, channelID, chaincodeName, chaincodeVersion, chaincodePath, codePackage, chaincodeCtorJSON, policy string) error {
	spec := &pb.ChaincodeSpec{}
	input := &pb.ChaincodeInput{}
	if err := json.Unmarshal([]byte(chaincodeCtorJSON), &input); err != nil {
		return err
	}
	spec = &pb.ChaincodeSpec{
		Type:        pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value[strings.ToUpper(ChaincodeLang)]),
		ChaincodeId: &pb.ChaincodeID{Path: chaincodePath, Name: chaincodeName, Version: chaincodeVersion},
		Input:       input,
	}
	codePackageBytes, err := ioutil.ReadFile(codePackage)
	if err != nil {
		return err
	}
	chaincodeDeploymentSpec := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: spec, CodePackage: codePackageBytes}

	creator, err := chainClient.Signer.Serialize()
	if err != nil {
		return err
	}

	p, err := cauthdsl.FromString(policy)
	if err != nil {
		return err
	}
	policyMarshalled := utils.MarshalOrPanic(p)

	prop, _, err := utils.CreateDeployProposalFromCDS(channelID, chaincodeDeploymentSpec, creator, policyMarshalled, []byte("escc"), []byte("vscc"), nil)
	if err != nil {
		return err
	}

	var signedProp *pb.SignedProposal
	signedProp, err = utils.GetSignedProposal(prop, chainClient.Signer)
	if err != nil {
		return err
	}

	// instantiate is currently only supported for one peer
	proposalResponse, err := chainClient.EndorserClient.ProcessProposal(context.Background(), signedProp)
	if err != nil {
		return err
	}

	if proposalResponse == nil {
		return errors.New("Error during instantiate: received nil proposal response")
	}

	// assemble a signed transaction (it's an Envelope message)
	env, err := utils.CreateSignedTx(prop, chainClient.Signer, proposalResponse)
	if err != nil {
		return err
	}

	if err = chainClient.BroadcastClient.Send(env); err != nil {
		return err
	}

	return nil
}

func UpgradeForChaincode(chainClient *ChainClient, channelID, chaincodeName, chaincodeVersion, chaincodePath, codePackage, chaincodeCtorJSON, policy string) error {
	spec := &pb.ChaincodeSpec{}
	input := &pb.ChaincodeInput{}
	if err := json.Unmarshal([]byte(chaincodeCtorJSON), &input); err != nil {
		return err
	}
	spec = &pb.ChaincodeSpec{
		Type:        pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value[strings.ToUpper(ChaincodeLang)]),
		ChaincodeId: &pb.ChaincodeID{Path: chaincodePath, Name: chaincodeName, Version: chaincodeVersion},
		Input:       input,
	}
	codePackageBytes, err := ioutil.ReadFile(codePackage)
	if err != nil {
		return err
	}
	chaincodeDeploymentSpec := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: spec, CodePackage: codePackageBytes}

	creator, err := chainClient.Signer.Serialize()
	if err != nil {
		return err
	}

	p, err := cauthdsl.FromString(policy)
	if err != nil {
		return err
	}
	policyMarshalled := utils.MarshalOrPanic(p)

	prop, _, err := utils.CreateUpgradeProposalFromCDS(channelID, chaincodeDeploymentSpec, creator, policyMarshalled, []byte("escc"), []byte("vscc"), nil)
	if err != nil {
		return err
	}

	var signedProp *pb.SignedProposal
	signedProp, err = utils.GetSignedProposal(prop, chainClient.Signer)
	if err != nil {
		return err
	}

	// instantiate is currently only supported for one peer
	proposalResponse, err := chainClient.EndorserClient.ProcessProposal(context.Background(), signedProp)
	if err != nil {
		return err
	}

	if proposalResponse == nil {
		return errors.New("Error during upgrade: received nil proposal response")
	}

	// assemble a signed transaction (it's an Envelope message)
	env, err := utils.CreateSignedTx(prop, chainClient.Signer, proposalResponse)
	if err != nil {
		return err
	}

	if err = chainClient.BroadcastClient.Send(env); err != nil {
		return err
	}

	return nil
}
