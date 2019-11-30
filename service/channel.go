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
