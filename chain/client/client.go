package client

import (
	"fmt"

	"github.com/hyperledger/fabric/msp"
	mspmgmt "github.com/hyperledger/fabric/msp/mgmt"
	"github.com/hyperledger/fabric/peer/common"
	peercom "github.com/hyperledger/fabric/peer/common"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/spf13/viper"
)

type ChainClient struct {
	Signer          msp.SigningIdentity
	EndorserClient  pb.EndorserClient
	BroadcastClient peercom.BroadcastClient
}

const (
	CHAINCODE_TYPE = "golang"
)

var (
	mspMgrConfigDir    string
	mspID              string
	mspType            string
	defaultChainClient *ChainClient
)

func InitConfig() error {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	mspMgrConfigDir = viper.GetString("peer.mspConfigPath")
	mspID = viper.GetString("peer.localMspId")
	mspType = msp.ProviderTypeToString(msp.FABRIC)

	return nil
}

func GetDefaultClient() (*ChainClient, error) {
	if defaultChainClient == nil {
		if len(mspMgrConfigDir) == 0 || len(mspID) == 0 {
			InitConfig()
		}

		var err error
		if defaultChainClient, err = NewClient(); err != nil {
			return nil, err
		}
	}
	return defaultChainClient, nil
}

func NewClient() (*ChainClient, error) {
	if len(mspMgrConfigDir) == 0 || len(mspID) == 0 {
		return nil, fmt.Errorf("run Init function first")
	}
	cc := new(ChainClient)
	err := peercom.InitCrypto(mspMgrConfigDir, mspID, mspType)
	if err != nil {
		return nil, err
	}
	if cc.Signer, err = mspmgmt.GetLocalMSP().GetDefaultSigningIdentity(); err != nil {
		return nil, err
	}
	if cc.EndorserClient, err = peercom.GetEndorserClient(peercom.UndefinedParamValue, peercom.UndefinedParamValue); err != nil {
		return nil, err
	}
	if cc.BroadcastClient, err = common.GetBroadcastClientFnc(); err != nil {
		return nil, err
	}
	return cc, nil
}
