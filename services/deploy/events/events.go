package events

import (
	"context"
	"fmt"
	"net"

	"github.com/kubemq-io/kubemq-go"
)

type Bus struct {
	Channel       string
	ClientId      string
	Context context.Context
	TransportType kubemq.TransportType
}

func (b *Bus) Connect() (kubemq.QueuesClient, error) {
	busIp, err := b.resolveBus()

	if err != nil {
		return kubemq.QueuesClient{}, err
	}

	client, err := kubemq.NewQueuesStreamClient(
		b.Context,
		kubemq.WithAddress(busIp, 50000),
		kubemq.WithClientId(b.ClientId),
		kubemq.WithTransportType(b.TransportType),
	)

	if err != nil {
		return kubemq.QueuesClient{}, err
	}

	return *client, err
}

func (b *Bus) Subscribe(client kubemq.QueuesClient, handler func(msg *kubemq.ReceiveQueueMessagesResponse, err error)) (chan struct{}, error) {
	fmt.Println("connected")

	done, subscribeErr := client.Subscribe(b.Context, &kubemq.ReceiveQueueMessagesRequest{
		Channel: b.Channel,
		ClientID: b.ClientId,
		MaxNumberOfMessages: 10,
		WaitTimeSeconds: 2,
	}, func(msgs *kubemq.ReceiveQueueMessagesResponse, err error) {
		handler(msgs, err)
	})

	if subscribeErr != nil {
		return nil, subscribeErr
	}

	return done, nil
}

func (b *Bus) resolveBus() (string, error) {
	dns := "kubemq-cluster-grpc.kubemq.svc.cluster.local"

	ip, err := net.LookupIP(dns)

	if err != nil {
		return "", err
	}

	return ip[0].String(), nil
}
