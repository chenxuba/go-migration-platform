package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	defaultNATSURL = "nats://127.0.0.1:4222"
	connectTimeout = 3 * time.Second
	adminTimeout   = 5 * time.Second
	defaultTag     = "default"
)

type Event struct {
	Topic string
	Tag   string
	Body  any
}

type Client struct {
	url        string
	group      string
	prefix     string
	streamName string
	conn       *nats.Conn
	js         jetstream.JetStream
}

type ConsumerHandler func(topic string, tag string, body []byte) error

type Consumer struct {
	url        string
	group      string
	prefix     string
	streamName string
	conn       *nats.Conn
	js         jetstream.JetStream

	mu            sync.Mutex
	subscriptions []subscription
	consumers     []jetstream.ConsumeContext
	started       bool
}

type subscription struct {
	topic   string
	tag     string
	handler ConsumerHandler
}

func NewClient(natsURL, group, prefix string) (*Client, error) {
	conn, js, streamName, err := connectJetStream(natsURL, group, prefix)
	if err != nil {
		return nil, err
	}
	return &Client{
		url:        normalizeURL(natsURL),
		group:      group,
		prefix:     prefix,
		streamName: streamName,
		conn:       conn,
		js:         js,
	}, nil
}

func (client *Client) Close() error {
	if client == nil || client.conn == nil {
		return nil
	}
	client.conn.Close()
	return nil
}

func (client *Client) Health() map[string]any {
	if client == nil || client.conn == nil {
		return map[string]any{"ok": false, "message": "not configured"}
	}
	if client.conn.IsConnected() {
		return map[string]any{
			"ok":      true,
			"url":     client.url,
			"group":   client.group,
			"stream":  client.streamName,
			"status":  client.conn.Status().String(),
			"message": fmt.Sprintf("connected to %s", client.url),
		}
	}
	address := tcpAddress(client.url)
	conn, err := net.DialTimeout("tcp", address, connectTimeout)
	if err != nil {
		return map[string]any{
			"ok":      false,
			"url":     client.url,
			"group":   client.group,
			"stream":  client.streamName,
			"status":  client.conn.Status().String(),
			"message": err.Error(),
		}
	}
	_ = conn.Close()
	return map[string]any{
		"ok":      false,
		"url":     client.url,
		"group":   client.group,
		"stream":  client.streamName,
		"status":  client.conn.Status().String(),
		"message": "tcp reachable but client is not connected",
	}
}

func (client *Client) Publish(ctx context.Context, event Event) error {
	if client == nil || client.js == nil {
		return fmt.Errorf("message publisher not ready")
	}
	payload, err := json.Marshal(event.Body)
	if err != nil {
		return err
	}
	msg := &nats.Msg{
		Subject: subjectFor(client.prefix, event.Topic, event.Tag),
		Header: nats.Header{
			"topic": []string{event.Topic},
			"tag":   []string{event.Tag},
		},
		Data: payload,
	}
	_, err = client.js.PublishMsg(ctx, msg)
	return err
}

func NewConsumer(natsURL, group, prefix string) (*Consumer, error) {
	conn, js, streamName, err := connectJetStream(natsURL, group, prefix)
	if err != nil {
		return nil, err
	}
	return &Consumer{
		url:        normalizeURL(natsURL),
		group:      group,
		prefix:     prefix,
		streamName: streamName,
		conn:       conn,
		js:         js,
	}, nil
}

func (consumerClient *Consumer) Subscribe(topic, tag string, handler ConsumerHandler) error {
	if consumerClient == nil {
		return fmt.Errorf("message consumer not ready")
	}
	if handler == nil {
		return fmt.Errorf("message handler is required")
	}
	consumerClient.mu.Lock()
	defer consumerClient.mu.Unlock()
	consumerClient.subscriptions = append(consumerClient.subscriptions, subscription{
		topic:   topic,
		tag:     tag,
		handler: handler,
	})
	if consumerClient.started {
		return consumerClient.startSubscription(subscription{topic: topic, tag: tag, handler: handler})
	}
	return nil
}

func (consumerClient *Consumer) Start() error {
	if consumerClient == nil {
		return fmt.Errorf("message consumer not ready")
	}
	consumerClient.mu.Lock()
	defer consumerClient.mu.Unlock()
	if consumerClient.started {
		return nil
	}
	for _, item := range consumerClient.subscriptions {
		if err := consumerClient.startSubscription(item); err != nil {
			return err
		}
	}
	consumerClient.started = true
	return nil
}

func (consumerClient *Consumer) Close() error {
	if consumerClient == nil {
		return nil
	}
	consumerClient.mu.Lock()
	defer consumerClient.mu.Unlock()
	for _, consumeCtx := range consumerClient.consumers {
		if consumeCtx != nil {
			consumeCtx.Stop()
		}
	}
	consumerClient.consumers = nil
	if consumerClient.conn != nil {
		consumerClient.conn.Close()
	}
	return nil
}

func (consumerClient *Consumer) startSubscription(item subscription) error {
	if consumerClient.js == nil {
		return fmt.Errorf("message consumer not ready")
	}
	ctx, cancel := context.WithTimeout(context.Background(), adminTimeout)
	defer cancel()
	durable := durableName(consumerClient.group, item.topic, item.tag)
	consumer, err := consumerClient.js.CreateOrUpdateConsumer(ctx, consumerClient.streamName, jetstream.ConsumerConfig{
		Durable:       durable,
		AckPolicy:     jetstream.AckExplicitPolicy,
		DeliverPolicy: jetstream.DeliverNewPolicy,
		FilterSubject: subjectPattern(consumerClient.prefix, item.topic, item.tag),
	})
	if err != nil {
		return err
	}
	consumeCtx, err := consumer.Consume(func(msg jetstream.Msg) {
		topic := msg.Headers().Get("topic")
		tag := msg.Headers().Get("tag")
		if topic == "" {
			topic, tag = topicAndTagFromSubject(msg.Subject())
		}
		if err := item.handler(topic, tag, msg.Data()); err != nil {
			_ = msg.Nak()
			return
		}
		_ = msg.Ack()
	})
	if err != nil {
		return err
	}
	consumerClient.consumers = append(consumerClient.consumers, consumeCtx)
	return nil
}

func connectJetStream(natsURL, group, prefix string) (*nats.Conn, jetstream.JetStream, string, error) {
	normalizedURL := normalizeURL(natsURL)
	conn, err := nats.Connect(normalizedURL, nats.Name(safeName(group, "go_migration_platform")), nats.Timeout(connectTimeout))
	if err != nil {
		return nil, nil, "", err
	}
	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		return nil, nil, "", err
	}
	streamName := streamNameFor(prefix)
	ctx, cancel := context.WithTimeout(context.Background(), adminTimeout)
	defer cancel()
	if _, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      streamName,
		Subjects:  []string{subjectPrefix(prefix) + ".>"},
		Retention: jetstream.LimitsPolicy,
		Storage:   jetstream.FileStorage,
	}); err != nil {
		conn.Close()
		return nil, nil, "", err
	}
	return conn, js, streamName, nil
}

func normalizeURL(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultNATSURL
	}
	return value
}

func tcpAddress(raw string) string {
	value := strings.Split(normalizeURL(raw), ",")[0]
	if !strings.Contains(value, "://") {
		value = "nats://" + value
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "127.0.0.1:4222"
	}
	host := parsed.Hostname()
	if host == "" {
		host = "127.0.0.1"
	}
	port := parsed.Port()
	if port == "" {
		port = "4222"
	}
	return net.JoinHostPort(host, port)
}

func subjectFor(prefix, topic, tag string) string {
	tagToken := subjectToken(tag, defaultTag)
	return subjectPrefix(prefix) + "." + subjectToken(topic, "event") + "." + tagToken
}

func subjectPattern(prefix, topic, tag string) string {
	if strings.TrimSpace(tag) == "" {
		return subjectPrefix(prefix) + "." + subjectToken(topic, "event") + ".*"
	}
	return subjectFor(prefix, topic, tag)
}

func subjectPrefix(prefix string) string {
	return "gmp." + subjectToken(prefix, "default")
}

func subjectToken(value, fallback string) string {
	token := sanitize(value, fallback, false)
	return strings.ToLower(token)
}

func streamNameFor(prefix string) string {
	return "GMP_" + safeName(prefix, "DEFAULT")
}

func durableName(parts ...string) string {
	name := safeName(strings.Join(parts, "_"), "CONSUMER")
	if len(name) > 120 {
		return name[:120]
	}
	return name
}

func safeName(value, fallback string) string {
	return strings.ToUpper(sanitize(value, fallback, true))
}

func sanitize(value, fallback string, upper bool) string {
	value = strings.TrimSpace(value)
	var builder strings.Builder
	lastUnderscore := false
	for _, item := range value {
		if item >= 'a' && item <= 'z' || item >= 'A' && item <= 'Z' || item >= '0' && item <= '9' {
			builder.WriteRune(item)
			lastUnderscore = false
			continue
		}
		if item == '_' || item == '-' {
			builder.WriteRune(item)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			builder.WriteByte('_')
			lastUnderscore = true
		}
	}
	result := strings.Trim(builder.String(), "_-")
	if result == "" {
		result = fallback
	}
	if upper {
		return strings.ToUpper(result)
	}
	return result
}

func topicAndTagFromSubject(subject string) (string, string) {
	parts := strings.Split(subject, ".")
	if len(parts) < 2 {
		return subject, ""
	}
	tag := parts[len(parts)-1]
	if tag == defaultTag {
		tag = ""
	}
	return parts[len(parts)-2], tag
}
