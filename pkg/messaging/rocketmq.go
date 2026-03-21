package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	rmq "github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

type RocketMQEvent struct {
	Topic string
	Tag   string
	Body  any
}

type RocketMQClient struct {
	nameServer string
	group      string
	prefix     string
	producer   rmq.Producer
}

type ConsumerHandler func(topic string, tag string, body []byte) error

type RocketMQConsumer struct {
	nameServer string
	group      string
	consumer   rmq.PushConsumer
	prefix     string
}

func NewRocketMQClient(nameServer, group, prefix string) (*RocketMQClient, error) {
	p, err := rmq.NewProducer(
		producer.WithNameServer([]string{nameServer}),
		producer.WithGroupName(group),
		producer.WithRetry(2),
	)
	if err != nil {
		return nil, err
	}
	if err := p.Start(); err != nil {
		return nil, err
	}
	return &RocketMQClient{
		nameServer: nameServer,
		group:      group,
		prefix:     prefix,
		producer:   p,
	}, nil
}

func (client *RocketMQClient) Close() error {
	if client == nil || client.producer == nil {
		return nil
	}
	return client.producer.Shutdown()
}

func (client *RocketMQClient) Health() map[string]any {
	conn, err := net.DialTimeout("tcp", client.nameServer, 3*time.Second)
	if err != nil {
		return map[string]any{
			"ok":         false,
			"nameServer": client.nameServer,
			"message":    err.Error(),
		}
	}
	_ = conn.Close()
	return map[string]any{
		"ok":         true,
		"nameServer": client.nameServer,
		"group":      client.group,
		"message":    fmt.Sprintf("connected to %s", client.nameServer),
	}
}

func (client *RocketMQClient) Publish(ctx context.Context, event RocketMQEvent) error {
	if client == nil || client.producer == nil {
		return fmt.Errorf("rocketmq producer not ready")
	}
	payload, err := json.Marshal(event.Body)
	if err != nil {
		return err
	}
	msg := &primitive.Message{
		Topic: client.prefixedTopic(event.Topic),
		Body:  payload,
	}
	if event.Tag != "" {
		msg.WithTag(event.Tag)
	}
	_, err = client.producer.SendSync(ctx, msg)
	return err
}

func (client *RocketMQClient) prefixedTopic(topic string) string {
	if client.prefix == "" {
		return topic
	}
	return client.prefix + "-" + topic
}

func NewRocketMQConsumer(nameServer, group, prefix string) (*RocketMQConsumer, error) {
	c, err := rmq.NewPushConsumer(
		consumer.WithNameServer([]string{nameServer}),
		consumer.WithGroupName(group),
	)
	if err != nil {
		return nil, err
	}
	return &RocketMQConsumer{
		nameServer: nameServer,
		group:      group,
		consumer:   c,
		prefix:     prefix,
	}, nil
}

func (consumerClient *RocketMQConsumer) Subscribe(topic, tag string, handler ConsumerHandler) error {
	selector := consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: "*",
	}
	if tag != "" {
		selector.Expression = tag
	}

	return consumerClient.consumer.Subscribe(consumerClient.prefixedTopic(topic), selector, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, msg := range msgs {
			if err := handler(msg.Topic, msg.GetTags(), msg.Body); err != nil {
				return consumer.ConsumeRetryLater, err
			}
		}
		return consumer.ConsumeSuccess, nil
	})
}

func (consumerClient *RocketMQConsumer) Start() error {
	return consumerClient.consumer.Start()
}

func (consumerClient *RocketMQConsumer) Close() error {
	if consumerClient == nil || consumerClient.consumer == nil {
		return nil
	}
	return consumerClient.consumer.Shutdown()
}

func (consumerClient *RocketMQConsumer) prefixedTopic(topic string) string {
	if consumerClient.prefix == "" {
		return topic
	}
	return consumerClient.prefix + "-" + topic
}
