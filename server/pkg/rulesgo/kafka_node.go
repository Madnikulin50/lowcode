package rulesgo

import (
	"context"
	"fmt"
	"strings"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

// KafkaConfig is the shared config shape for kafka.produce, kafka.consume
// and kafka.subscribe nodes. Not every field applies to every node type -
// see each executor for which ones it reads.
type KafkaConfig struct {
	Brokers string `json:"brokers"` // comma-separated host:port list
	Topic   string `json:"topic"`
	GroupID string `json:"groupId,omitempty"`

	// kafka.produce only
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`

	// kafka.consume only
	MaxMessages    int `json:"maxMessages,omitempty"`
	TimeoutSeconds int `json:"timeoutSeconds,omitempty"`

	// kafka.subscribe only
	IngestChainID string `json:"ingestChainID,omitempty"`
}

func (c KafkaConfig) brokerList() []string {
	var out []string
	for _, b := range strings.Split(c.Brokers, ",") {
		b = strings.TrimSpace(b)
		if b != "" {
			out = append(out, b)
		}
	}
	return out
}

// kafkaWriter/kafkaReader are the minimal surface used from *kafka.Writer and
// *kafka.Reader. Keeping them as small interfaces (rather than depending on
// the concrete types directly) lets tests substitute a fake broker instead
// of requiring a live Kafka cluster.
type kafkaWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type kafkaReader interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

var newKafkaWriter = func(cfg KafkaConfig) kafkaWriter {
	return &kafka.Writer{
		Addr:                   kafka.TCP(cfg.brokerList()...),
		Topic:                  cfg.Topic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
}

var newKafkaReader = func(cfg KafkaConfig) kafkaReader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.brokerList(),
		Topic:   cfg.Topic,
		GroupID: cfg.GroupID,
	})
}

// --- kafka.produce ---

type kafkaProduceExecutor struct{}

func (n *kafkaProduceExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	cfg, err := ParseNodeConfig[KafkaConfig](node.Config)
	if err != nil {
		return nil, err
	}
	cfg.Brokers = resolveTemplateValue(cfg.Brokers, ec)
	cfg.Topic = resolveTemplateValue(cfg.Topic, ec)
	key := resolveTemplateValue(cfg.Key, ec)
	value := resolveTemplateValue(cfg.Value, ec)

	if len(cfg.brokerList()) == 0 {
		return nil, fmt.Errorf("kafka.produce: brokers is required")
	}
	if cfg.Topic == "" {
		return nil, fmt.Errorf("kafka.produce: topic is required")
	}

	w := newKafkaWriter(cfg)
	defer w.Close()

	msg := kafka.Message{Topic: cfg.Topic, Value: []byte(value)}
	if key != "" {
		msg.Key = []byte(key)
	}

	writeCtx := ctx
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		writeCtx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	if err := w.WriteMessages(writeCtx, msg); err != nil {
		return nil, fmt.Errorf("kafka.produce: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"topic":   cfg.Topic,
		"key":     key,
	}, nil
}

// --- kafka.consume ---
//
// Reads up to MaxMessages messages (default 10) within TimeoutSeconds
// (default 5), committing offsets for whatever was read, and returns them as
// an array so the rest of the chain can `foreach` over it. This is the
// "batch fetch" action - for continuous, event-driven consumption see
// kafka.subscribe.

type kafkaConsumeExecutor struct{}

func (n *kafkaConsumeExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	cfg, err := ParseNodeConfig[KafkaConfig](node.Config)
	if err != nil {
		return nil, err
	}
	cfg.Brokers = resolveTemplateValue(cfg.Brokers, ec)
	cfg.Topic = resolveTemplateValue(cfg.Topic, ec)
	cfg.GroupID = resolveTemplateValue(cfg.GroupID, ec)

	if len(cfg.brokerList()) == 0 {
		return nil, fmt.Errorf("kafka.consume: brokers is required")
	}
	if cfg.Topic == "" {
		return nil, fmt.Errorf("kafka.consume: topic is required")
	}
	if cfg.GroupID == "" {
		cfg.GroupID = "lowcode-rulechain"
	}
	maxMessages := cfg.MaxMessages
	if maxMessages <= 0 {
		maxMessages = 10
	}
	timeout := cfg.TimeoutSeconds
	if timeout <= 0 {
		timeout = 5
	}

	r := newKafkaReader(cfg)
	defer r.Close()

	readCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	messages := make([]map[string]interface{}, 0, maxMessages)
	committed := make([]kafka.Message, 0, maxMessages)
	for len(messages) < maxMessages {
		m, err := r.FetchMessage(readCtx)
		if err != nil {
			// timeout or no more messages waiting - not fatal, return what we have
			break
		}
		messages = append(messages, map[string]interface{}{
			"key":       string(m.Key),
			"value":     string(m.Value),
			"partition": m.Partition,
			"offset":    m.Offset,
		})
		committed = append(committed, m)
	}

	if len(committed) > 0 {
		commitCtx, cancelCommit := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelCommit()
		if err := r.CommitMessages(commitCtx, committed...); err != nil {
			return nil, fmt.Errorf("kafka.consume: commit failed: %w", err)
		}
	}

	return map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
	}, nil
}

// --- kafka.subscribe ---
//
// Non-blocking: starts a background consumer (via the injected start func,
// normally BrokerSubscriber.StartKafkaSubscribe) that feeds an ingest chain
// for every message received, and returns immediately - the same shape as
// the existing `detach` node.

type kafkaSubscribeExecutor struct {
	start func(ctx context.Context, subKey string, cfg KafkaConfig, ingestChainID string) error
}

func (n *kafkaSubscribeExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	cfg, err := ParseNodeConfig[KafkaConfig](node.Config)
	if err != nil {
		return nil, err
	}
	cfg.Brokers = resolveTemplateValue(cfg.Brokers, ec)
	cfg.Topic = resolveTemplateValue(cfg.Topic, ec)
	cfg.GroupID = resolveTemplateValue(cfg.GroupID, ec)
	cfg.IngestChainID = resolveTemplateValue(cfg.IngestChainID, ec)

	if len(cfg.brokerList()) == 0 {
		return nil, fmt.Errorf("kafka.subscribe: brokers is required")
	}
	if cfg.Topic == "" {
		return nil, fmt.Errorf("kafka.subscribe: topic is required")
	}
	if cfg.IngestChainID == "" {
		return nil, fmt.Errorf("kafka.subscribe: ingestChainID is required")
	}

	if n.start == nil {
		return map[string]interface{}{"status": "kafka_subscribe_not_configured"}, nil
	}

	subKey := node.ID + ":" + cfg.Topic
	if err := n.start(ctx, subKey, cfg, cfg.IngestChainID); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"subscribed":    true,
		"topic":         cfg.Topic,
		"ingestChainID": cfg.IngestChainID,
	}, nil
}
