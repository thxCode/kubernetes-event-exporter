package client

import (
	"github.com/rancher/norman/clientbase"
)

type Client struct {
	clientbase.APIBaseClient

	HuaWeiClusterEvent HuaWeiClusterEventOperations
	Pod                PodOperations
	Node               NodeOperations
	HuaWeiEventLog     HuaWeiEventLogOperations
}

func NewClient(opts *clientbase.ClientOpts) (*Client, error) {
	baseClient, err := clientbase.NewAPIClient(opts)
	if err != nil {
		return nil, err
	}

	client := &Client{
		APIBaseClient: baseClient,
	}

	client.HuaWeiClusterEvent = newHuaWeiClusterEventClient(client)
	client.Pod = newPodClient(client)
	client.Node = newNodeClient(client)
	client.HuaWeiEventLog = newHuaWeiEventLogClient(client)

	return client, nil
}
