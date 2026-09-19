package rulesgo

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// BrokerSubscriber runs background message-broker consumers (Kafka,
// RabbitMQ) that feed an ingest chain for every message received - the
// "trigger" half of broker integration, complementing the synchronous
// kafka.produce/consume and rabbitmq.publish/consume action nodes.
//
// It mirrors AgentPoller (used by the `detach` node): non-blocking start,
// one goroutine per subscription, keyed so re-running the chain that owns
// the node replaces rather than piles up subscriptions.
type BrokerSubscriber struct {
	engine *EngineWithPersistence

	mu      sync.Mutex
	cancels map[string]context.CancelFunc
}

func NewBrokerSubscriber() *BrokerSubscriber {
	return &BrokerSubscriber{cancels: make(map[string]context.CancelFunc)}
}

func (b *BrokerSubscriber) SetEngine(e *EngineWithPersistence) {
	b.engine = e
}

// Stop cancels a running subscription started under subKey, if any.
func (b *BrokerSubscriber) Stop(subKey string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if cancel, ok := b.cancels[subKey]; ok {
		cancel()
		delete(b.cancels, subKey)
	}
}

// replace registers a new cancel func under subKey, cancelling whatever was
// previously registered under the same key first.
func (b *BrokerSubscriber) replace(subKey string, cancel context.CancelFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if prev, ok := b.cancels[subKey]; ok {
		prev()
	}
	b.cancels[subKey] = cancel
}

func (b *BrokerSubscriber) runIngest(ingestChainID string, envelope map[string]interface{}, triggerType string) {
	if b.engine == nil {
		log.Printf("[rulesgo] broker subscriber: engine not configured, dropping message for %s", ingestChainID)
		return
	}
	defer func() {
		if p := recover(); p != nil {
			log.Printf("[rulesgo] broker subscriber: ingest %s panic: %v", ingestChainID, p)
		}
	}()
	if _, err := b.engine.RunWithLog(context.Background(), ingestChainID, envelope, triggerType); err != nil {
		log.Printf("[rulesgo] broker subscriber: ingest %s: %v", ingestChainID, err)
	}
}

// StartKafkaSubscribe starts (or restarts) a background Kafka consumer for
// subKey. Non-blocking: launches a goroutine and returns immediately.
func (b *BrokerSubscriber) StartKafkaSubscribe(ctx context.Context, subKey string, cfg KafkaConfig, ingestChainID string) error {
	if len(cfg.brokerList()) == 0 {
		return fmt.Errorf("kafka.subscribe: brokers is required")
	}
	if cfg.Topic == "" {
		return fmt.Errorf("kafka.subscribe: topic is required")
	}
	if ingestChainID == "" {
		return fmt.Errorf("kafka.subscribe: ingestChainID is required")
	}
	if cfg.GroupID == "" {
		cfg.GroupID = "lowcode-rulechain"
	}

	runCtx, cancel := context.WithCancel(context.Background())
	b.replace(subKey, cancel)

	r := newKafkaReader(cfg)
	go func() {
		defer r.Close()
		for {
			if runCtx.Err() != nil {
				return
			}
			m, err := r.FetchMessage(runCtx)
			if err != nil {
				if runCtx.Err() != nil {
					return
				}
				log.Printf("[rulesgo] kafka.subscribe %s: %v", cfg.Topic, err)
				// back off briefly so a persistent broker error doesn't spin the CPU
				select {
				case <-runCtx.Done():
					return
				case <-time.After(2 * time.Second):
				}
				continue
			}

			envelope := NormalizeIngestEnvelope(map[string]interface{}{
				"topic":     m.Topic,
				"key":       string(m.Key),
				"value":     string(m.Value),
				"partition": m.Partition,
				"offset":    m.Offset,
			})
			b.runIngest(ingestChainID, envelope, "kafka-subscribe")

			if err := r.CommitMessages(runCtx, m); err != nil && runCtx.Err() == nil {
				log.Printf("[rulesgo] kafka.subscribe commit %s: %v", cfg.Topic, err)
			}
		}
	}()

	return nil
}

// StartRabbitMQSubscribe starts (or restarts) a background RabbitMQ consumer
// for subKey, feeding ingestChainID for every message received.
func (b *BrokerSubscriber) StartRabbitMQSubscribe(ctx context.Context, subKey string, cfg RabbitMQConfig, ingestChainID string) error {
	if cfg.URL == "" {
		return fmt.Errorf("rabbitmq.subscribe: url is required")
	}
	if cfg.Queue == "" {
		return fmt.Errorf("rabbitmq.subscribe: queue is required")
	}
	if ingestChainID == "" {
		return fmt.Errorf("rabbitmq.subscribe: ingestChainID is required")
	}

	conn, err := dialRabbitMQ(cfg.URL)
	if err != nil {
		return fmt.Errorf("rabbitmq.subscribe: dial: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("rabbitmq.subscribe: channel: %w", err)
	}
	deliveries, err := ch.Consume(cfg.Queue, "", false, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("rabbitmq.subscribe: %w", err)
	}

	runCtx, cancel := context.WithCancel(context.Background())
	b.replace(subKey, cancel)

	go func() {
		defer conn.Close()
		defer ch.Close()
		for {
			select {
			case <-runCtx.Done():
				return
			case d, ok := <-deliveries:
				if !ok {
					return
				}
				envelope := NormalizeIngestEnvelope(map[string]interface{}{
					"queue":      cfg.Queue,
					"routingKey": d.RoutingKey,
					"body":       string(d.Body),
				})
				b.runIngest(ingestChainID, envelope, "rabbitmq-subscribe")
				_ = d.Ack(false)
			}
		}
	}()

	return nil
}
