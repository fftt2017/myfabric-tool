package service

import (
	"myfabric-tool/chain/client"
	"myfabric-tool/model"
)

func ListInstalledChaincode() ([]*model.Chaincode, error) {
	chainClient, err := client.GetDefaultClient()
	if err != nil {
		return nil, err
	}
	installed, err := client.ListInstalledForChaincode(chainClient)
	if err != nil {
		return nil, err
	}

	installedChaincode := make([]*model.Chaincode, len(installed), len(installed))
	for i := 0; i < len(installed); i++ {
		installedChaincode[i] = model.NewChaincode(installed[i])
	}

	return installedChaincode, nil
}

func ListInstantiateChaincode(channelID string)([]*model.Chaincode, error){
	chainClient, err := client.GetDefaultClient()
	if err != nil {
		return nil, err
	}
	instantiated,err := client.ListInstantiatedForChaincode(chainClient,channelID)
	if err != nil {
		return nil, err
	}
	instantiatedChaincode := make([]*model.Chaincode, len(instantiated), len(instantiated))
	for i := 0; i < len(instantiated); i++ {
		instantiatedChaincode[i] = model.NewChaincode(instantiated[i])
	}
	return instantiatedChaincode, nil
}

func ChaincodeQuery(channelID string,chaincodeName string,args string) (string, error) {
	chainClient, err := client.GetDefaultClient()
	if err != nil {
		return "", err
	}
	result, err := client.QueryForChaincode(chainClient, channelID,chaincodeName,args)
	if err != nil {
		return "", err
	}
	return result, nil
}

func ChaincodeInvoke(channelID string, chaincodeName string, args string) (error) {
	chainClient, err := client.GetDefaultClient()
	if err != nil {
		return err
	}
	err = client.InvokeForChainCode(chainClient, channelID,chaincodeName,args)
	return err
}
