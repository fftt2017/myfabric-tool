package service

import (
	"myfabric-tool/chain/client"
	"myfabric-tool/model"
)

func ListChannels() ([]*model.Channel, error) {
	chainClient, err := client.GetDefaultClient()
	if err != nil {
		return nil, err
	}
	channelInfos, err := client.ListChannels(chainClient)
	if err != nil {
		return nil, err
	}

	channels := make([]*model.Channel, len(channelInfos), len(channelInfos))
	for i := 0; i < len(channelInfos); i++ {
		channels[i] = model.NewChannel(channelInfos[i])
	}

	return channels, nil
}

func GetChannelInfo(channelID string) (*model.Blockchain, error) {
	chainClient, err := client.GetDefaultClient()
	if err != nil {
		return nil, err
	}
	blockchainInfo, err := client.GetChannelInfo(chainClient, channelID)
	if err != nil {
		return nil, err
	}
	blockchain := model.NewBlockchain(blockchainInfo)

	return blockchain, nil
}

func ChannelFetch(channelID string, blockIndex string) (string, error) {
	chainClient, err := client.GetDefaultClient()
	if err != nil {
		return "", err
	}
	block, err := client.FetchBlock(chainClient, channelID,blockIndex)
	if err != nil {
		return "", err
	}


	return block, nil
}
