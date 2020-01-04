package client

import (
	"fmt"
	"path/filepath"

	"myfabric-tool/config"

	"github.com/hyperledger/fabric/bccsp/factory"
	"github.com/hyperledger/fabric/common/viperutil"
	"github.com/hyperledger/fabric/msp"
	mspmgmt "github.com/hyperledger/fabric/msp/mgmt"
	"github.com/hyperledger/fabric/peer/common"
	peercom "github.com/hyperledger/fabric/peer/common"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type ChainClient struct {
	Signer          msp.SigningIdentity
	EndorserClient  pb.EndorserClient
	BroadcastClient peercom.BroadcastClient
}

var (
	defaultChainClient *ChainClient
	curOrdererOrg      string
	curOrdererNode     string
	curPeerOrg         string
	curPeerNode        string
	curUserOrg         string
	curUser            string
)

func GetDefaultClient() (*ChainClient, error) {
	if defaultChainClient == nil {
		return nil, errors.New("default chain client is nil")
	}
	return defaultChainClient, nil
}

func RefreshDefaultClient(ordererOrg, ordererNode, peerOrg, peerNode, userOrg, user string) error {
	var err error
	if err = reloadEnv(ordererOrg, ordererNode, peerOrg, peerNode, userOrg, user); err != nil {
		return err
	}

	cc := new(ChainClient)
	if cc.Signer, err = initSigner(); err != nil {
		return err
	}

	if cc.EndorserClient, err = peercom.GetEndorserClient(peercom.UndefinedParamValue, peercom.UndefinedParamValue); err != nil {
		return err
	}

	if cc.BroadcastClient, err = common.GetBroadcastClientFnc(); err != nil {
		return err
	}

	return nil
}

func initSigner() (msp.SigningIdentity, error) {
	mspMgrConfigDir := viper.GetString("peer.mspConfigPath")
	mspID := viper.GetString("peer.localMspId")
	mspType := msp.ProviderTypeToString(msp.FABRIC)

	fmt.Println("peer.mspConfigPath:", mspMgrConfigDir)
	fmt.Println("peer.localMspId:", mspID)

	// Clean local msp
	mspmgmt.CleanLocalMSP()

	// Init the BCCSP
	peercom.SetBCCSPKeystorePath()
	var bccspConfig *factory.FactoryOpts
	if err := viperutil.EnhancedExactUnmarshalKey("peer.BCCSP", &bccspConfig); err != nil {
		return nil, err
	}
	keystoreDir := filepath.Join(mspMgrConfigDir, "keystore")
	bccspConfig = msp.SetupBCCSPKeystoreConfig(bccspConfig, keystoreDir)

	if err := factory.InitFactoriesAgain(bccspConfig); err != nil {
		return nil, err
	}

	if err := peercom.InitCrypto(mspMgrConfigDir, mspID, mspType); err != nil {
		return nil, err
	}
	return mspmgmt.GetLocalMSP().GetDefaultSigningIdentity()
}

func reloadEnv(ordererOrg, ordererNode, peerOrg, peerNode, userOrg, user string) error {
	if curOrdererOrg != ordererOrg || curOrdererNode != ordererNode {
		curOrdererOrg = ordererOrg
		curOrdererNode = ordererNode
		if err := setOrdererEnv(curOrdererOrg, curOrdererNode); err != nil {
			return err
		}
	}

	if curPeerOrg != peerOrg || curPeerNode != peerNode {
		curPeerOrg = peerOrg
		curPeerNode = peerNode
		if err := setPeerEnv(curPeerOrg, curPeerNode); err != nil {
			return err
		}
	}

	if curUserOrg != userOrg || curUser != user {
		curUserOrg = userOrg
		curUser = user
		if err := setUserEnv(curUserOrg, curUser); err != nil {
			return err
		}
	}

	return nil
}

func setOrdererEnv(orgstr, nodestr string) error {
	org := config.GetOrg(orgstr)
	if org == nil {
		return fmt.Errorf("can't find org %s", orgstr)
	}

	node := org.GetNode(nodestr)
	if node == nil {
		return fmt.Errorf("can't find org %s node %s", orgstr, nodestr)
	}
	if node.Type != config.TYPE_ORDERER {
		return fmt.Errorf("org %s node %s type isn't orderer", orgstr, nodestr)
	}

	viper.Set("orderer.address", node.Address)
	viper.Set("orderer.tls.serverhostoverride", node.TLSServerHostOverride)
	if node.TLSEnabled {
		viper.Set("orderer.tls.enabled", true)
		viper.Set("orderer.tls.rootcert.file", org.TLSRootCertFile)

	}
	if node.TLSClientAuthRequired {
		viper.Set("orderer.tls.clientAuthRequired", true)
	}

	return nil
}

func setPeerEnv(orgstr, nodestr string) error {
	org := config.GetOrg(orgstr)
	if org == nil {
		return fmt.Errorf("can't find org %s", orgstr)
	}

	node := org.GetNode(nodestr)
	if node == nil {
		return fmt.Errorf("can't find org %s node %s", orgstr, nodestr)
	}
	if node.Type != config.TYPE_PEER {
		return fmt.Errorf("org %s node %s type isn't peer", orgstr, nodestr)
	}

	viper.Set("peer.address", node.Address)
	viper.Set("peer.tls.serverhostoverride", node.TLSServerHostOverride)
	if node.TLSEnabled {
		viper.Set("peer.tls.enabled", true)
		viper.Set("peer.tls.rootcert.file", org.TLSRootCertFile)
	}
	if node.TLSClientAuthRequired {
		viper.Set("peer.tls.clientAuthRequired", true)
	}

	return nil
}

func setUserEnv(orgstr, userstr string) error {
	org := config.GetOrg(orgstr)
	if org == nil {
		return fmt.Errorf("can't find org %s", orgstr)
	}

	user := org.GetUser(userstr)
	if user == nil {
		return fmt.Errorf("can't find org %s user %s", orgstr, userstr)
	}

	viper.Set("peer.localMspId", org.LocalMSPId)
	viper.Set("peer.mspConfigPath", user.MspConfigPath)

	viper.Set("peer.tls.clientKey.file", user.TLSClientKeyFile)
	viper.Set("peer.tls.clientCert.file", user.TLSClientCertFile)
	viper.Set("orderer.tls.clientKey.file", user.TLSClientKeyFile)
	viper.Set("orderer.tls.clientCert.file", user.TLSClientCertFile)

	return nil
}
