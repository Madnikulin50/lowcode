package rulesgo

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/madnikulin50/lowcode/server/pkg/vault"
)

// RabbitMQConfig is the shared config shape for rabbitmq.publish,
// rabbitmq.consume and rabbitmq.subscribe nodes.
type RabbitMQConfig struct {
	URL string `json:"url"`
	// URLSecretRef points URL at a Vault secret instead of storing the AMQP
	// URL (which usually embeds credentials, e.g. amqp://user:pass@host) in
	// plaintext. When set, it always wins over URL - see resolveAMQPURL.
	URLSecretRef string `json:"urlSecretRef,omitempty"`
	Exchange     string `json:"exchange,omitempty"`
	RoutingKey   string `json:"routingKey,omitempty"`
	Queue        string `json:"queue,omitempty"`

	// rabbitmq.publish only
	Body string `json:"body,omitempty"`

	// rabbitmq.consume only
	MaxMessages    int `json:"maxMessages,omitempty"`
	TimeoutSeconds int `json:"timeoutSeconds,omitempty"`

	// rabbitmq.subscribe only
	IngestChainID string `json:"ingestChainID,omitempty"`
}

// rabbitmqChannel/rabbitmqConnection are the minimal surface used from
// *amqp.Channel / *amqp.Connection. Small interfaces so tests can substitute
// a fake broker instead of requiring a live RabbitMQ server.
type rabbitmqChannel interface {
	PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	Close() error
}

type rabbitmqConnection interface {
	Channel() (rabbitmqChannel, error)
	Close() error
}

type amqpConnWrapper struct{ conn *amqp.Connection }

func (w *amqpConnWrapper) Channel() (rabbitmqChannel, error) {
	return w.conn.Channel()
}
func (w *amqpConnWrapper) Close() error { return w.conn.Close() }

var dialRabbitMQ = func(url string) (rabbitmqConnection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	return &amqpConnWrapper{conn: conn}, nil
}

// resolveAMQPURL resolves templates on URL/URLSecretRef, then - if
// URLSecretRef is set - resolves the actual URL from Vault, overriding the
// plaintext URL field. This is the only place secret resolution happens, so
// every node (publish/consume/subscribe) gets it for free.
func resolveAMQPURL(ctx context.Context, cfg *RabbitMQConfig, ec *ExecutionContext) error {
	cfg.URL = resolveTemplateValue(cfg.URL, ec)
	ref := resolveTemplateValue(cfg.URLSecretRef, ec)
	resolved, err := vault.ResolveField(ctx, ref, cfg.URL)
	if err != nil {
		return fmt.Errorf("resolve AMQP url: %w", err)
	}
	cfg.URL = resolved
	return nil
}

// --- rabbitmq.publish ---

type rabbitmqPublishExecutor struct{}

func (n *rabbitmqPublishExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	cfg, err := ParseNodeConfig[RabbitMQConfig](node.Config)
	if err != nil {
		return nil, err
	}
	if err := resolveAMQPURL(ctx, &cfg, ec); err != nil {
		return nil, fmt.Errorf("rabbitmq.publish: %w", err)
	}
	cfg.Exchange = resolveTemplateValue(cfg.Exchange, ec)
	cfg.RoutingKey = resolveTemplateValue(cfg.RoutingKey, ec)
	cfg.Queue = resolveTemplateValue(cfg.Queue, ec)
	body := resolveTemplateValue(cfg.Body, ec)

	if cfg.URL == "" {
		return nil, fmt.Errorf("rabbitmq.publish: url is required")
	}
	routingKey := cfg.RoutingKey
	if routingKey == "" {
		routingKey = cfg.Queue
	}
	if routingKey == "" {
		return nil, fmt.Errorf("rabbitmq.publish: routingKey or queue is required")
	}

	conn, err := dialRabbitMQ(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq.publish: dial: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("rabbitmq.publish: channel: %w", err)
	}
	defer ch.Close()

	pubCtx := ctx
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		pubCtx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	err = ch.PublishWithContext(pubCtx, cfg.Exchange, routingKey, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(body),
	})
	if err != nil {
		return nil, fmt.Errorf("rabbitmq.publish: %w", err)
	}

	return map[string]interface{}{
		"success":    true,
		"exchange":   cfg.Exchange,
		"routingKey": routingKey,
	}, nil
}

// --- rabbitmq.consume ---
//
// Reads up to MaxMessages messages (default 10) within TimeoutSeconds
// (default 5), acking each, and returns them as an array. This is the "batch
// fetch" action - for continuous, event-driven consumption see
// rabbitmq.subscribe.

type rabbitmqConsumeExecutor struct{}

func (n *rabbitmqConsumeExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	cfg, err := ParseNodeConfig[RabbitMQConfig](node.Config)
	if err != nil {
		return nil, err
	}
	if err := resolveAMQPURL(ctx, &cfg, ec); err != nil {
		return nil, fmt.Errorf("rabbitmq.consume: %w", err)
	}
	cfg.Queue = resolveTemplateValue(cfg.Queue, ec)

	if cfg.URL == "" {
		return nil, fmt.Errorf("rabbitmq.consume: url is required")
	}
	if cfg.Queue == "" {
		return nil, fmt.Errorf("rabbitmq.consume: queue is required")
	}

	maxMessages := cfg.MaxMessages
	if maxMessages <= 0 {
		maxMessages = 10
	}
	timeout := cfg.TimeoutSeconds
	if timeout <= 0 {
		timeout = 5
	}

	conn, err := dialRabbitMQ(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq.consume: dial: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("rabbitmq.consume: channel: %w", err)
	}
	defer ch.Close()

	deliveries, err := ch.Consume(cfg.Queue, "", false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq.consume: %w", err)
	}

	messages := make([]map[string]interface{}, 0, maxMessages)
	timer := time.NewTimer(time.Duration(timeout) * time.Second)
	defer timer.Stop()

loop:
	for len(messages) < maxMessages {
		select {
		case d, ok := <-deliveries:
			if !ok {
				break loop
			}
			messages = append(messages, map[string]interface{}{
				"body":        string(d.Body),
				"routingKey":  d.RoutingKey,
				"deliveryTag": d.DeliveryTag,
			})
			_ = d.Ack(false)
		case <-timer.C:
			break loop
		case <-ctx.Done():
			break loop
		}
	}

	return map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
	}, nil
}

// --- rabbitmq.subscribe ---
//
// Non-blocking: starts a background consumer (via the injected start func,
// normally BrokerSubscriber.StartRabbitMQSubscribe) that feeds an ingest
// chain for every message received, and returns immediately.

type rabbitmqSubscribeExecutor struct {
	start func(ctx context.Context, subKey string, cfg RabbitMQConfig, ingestChainID string) error
}

func (n *rabbitmqSubscribeExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	cfg, err := ParseNodeConfig[RabbitMQConfig](node.Config)
	if err != nil {
		return nil, err
	}
	if err := resolveAMQPURL(ctx, &cfg, ec); err != nil {
		return nil, fmt.Errorf("rabbitmq.subscribe: %w", err)
	}
	cfg.Queue = resolveTemplateValue(cfg.Queue, ec)
	cfg.IngestChainID = resolveTemplateValue(cfg.IngestChainID, ec)

	if cfg.URL == "" {
		return nil, fmt.Errorf("rabbitmq.subscribe: url is required")
	}
	if cfg.Queue == "" {
		return nil, fmt.Errorf("rabbitmq.subscribe: queue is required")
	}
	if cfg.IngestChainID == "" {
		return nil, fmt.Errorf("rabbitmq.subscribe: ingestChainID is required")
	}

	if n.start == nil {
		return map[string]interface{}{"status": "rabbitmq_subscribe_not_configured"}, nil
	}

	subKey := node.ID + ":" + cfg.Queue
	if err := n.start(ctx, subKey, cfg, cfg.IngestChainID); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"subscribed":    true,
		"queue":         cfg.Queue,
		"ingestChainID": cfg.IngestChainID,
	}, nil
}
