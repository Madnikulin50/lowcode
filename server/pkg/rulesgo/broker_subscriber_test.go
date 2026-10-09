package rulesgo

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	kafka "github.com/segmentio/kafka-go"
)

// newTestIngestEngine builds a minimal EngineWithPersistence with one chain
// ("ingest_test") that always succeeds, so we can assert on
// engine.ExecutionLogs() without needing real broker infrastructure.
func newTestIngestEngine(t *testing.T) *EngineWithPersistence {
	t.Helper()
	engine := NewEngineWithPersistence(DefaultRegistry(&DefaultConfig{}), NewMemoryPersistence())
	engine.RegisterChain(&Chain{
		ID:        "ingest_test",
		Name:      "Ingest Test",
		EntryNode: "n1",
		Nodes: []ChainNode{
			{ID: "n1", Type: "condition", Config: json.RawMessage(`{"field":"nope","operator":"empty"}`)},
		},
	})
	return engine
}

func waitForRuns(t *testing.T, engine *EngineWithPersistence, want int) []ExecRecord {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if logs := engine.ExecutionLogs(0); len(logs) >= want {
			return logs
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %d run(s), got %d", want, len(engine.ExecutionLogs(0)))
	return nil
}

func TestBrokerSubscriber_KafkaFeedsIngestChain(t *testing.T) {
	r := &fakeKafkaReader{queue: []kafka.Message{
		{Topic: "orders", Key: []byte("k1"), Value: []byte("v1")},
		{Topic: "orders", Key: []byte("k2"), Value: []byte("v2")},
	}}
	withFakeKafkaReader(t, r)

	engine := newTestIngestEngine(t)
	sub := NewBrokerSubscriber()
	sub.SetEngine(engine)

	err := sub.StartKafkaSubscribe(context.Background(), "sub1", KafkaConfig{
		Brokers: "b:9092",
		Topic:   "orders",
	}, "ingest_test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer sub.Stop("sub1")

	logs := waitForRuns(t, engine, 2)
	for _, l := range logs {
		if l.TriggerType != "kafka-subscribe" {
			t.Fatalf("unexpected triggerType: %s", l.TriggerType)
		}
		if l.Input["topic"] != "orders" {
			t.Fatalf("expected topic in envelope, got %#v", l.Input)
		}
	}
	if len(r.committed) != 2 {
		t.Fatalf("expected 2 committed messages, got %d", len(r.committed))
	}
}

func TestBrokerSubscriber_KafkaRestartsReplacesPreviousSubscription(t *testing.T) {
	r1 := &fakeKafkaReader{queue: []kafka.Message{{Topic: "t1"}}}
	withFakeKafkaReader(t, r1)

	engine := newTestIngestEngine(t)
	sub := NewBrokerSubscriber()
	sub.SetEngine(engine)

	if err := sub.StartKafkaSubscribe(context.Background(), "same-key", KafkaConfig{Brokers: "b:9092", Topic: "t1"}, "ingest_test"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	waitForRuns(t, engine, 1)

	// starting again under the same subKey must cancel the first goroutine
	// (and not leave it running) before starting the second.
	r2 := &fakeKafkaReader{queue: []kafka.Message{{Topic: "t2"}, {Topic: "t2"}}}
	withFakeKafkaReader(t, r2)
	if err := sub.StartKafkaSubscribe(context.Background(), "same-key", KafkaConfig{Brokers: "b:9092", Topic: "t2"}, "ingest_test"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	waitForRuns(t, engine, 3) // 1 from r1 + 2 from r2
	sub.Stop("same-key")

	if !r1.closed {
		t.Fatal("first subscription's reader should have been closed on replace")
	}
}

func TestBrokerSubscriber_RabbitMQFeedsIngestChain(t *testing.T) {
	ack := &fakeAcknowledger{}
	deliveries := make(chan amqp.Delivery, 2)
	deliveries <- amqp.Delivery{Body: []byte("m1"), RoutingKey: "rk", DeliveryTag: 1, Acknowledger: ack}
	deliveries <- amqp.Delivery{Body: []byte("m2"), RoutingKey: "rk", DeliveryTag: 2, Acknowledger: ack}

	ch := &fakeRabbitMQChannel{deliveries: deliveries}
	conn := &fakeRabbitMQConnection{ch: ch}
	withFakeRabbitMQDialer(t, conn)

	engine := newTestIngestEngine(t)
	sub := NewBrokerSubscriber()
	sub.SetEngine(engine)

	err := sub.StartRabbitMQSubscribe(context.Background(), "sub1", RabbitMQConfig{
		URL:   "amqp://x",
		Queue: "jobs",
	}, "ingest_test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer sub.Stop("sub1")

	logs := waitForRuns(t, engine, 2)
	for _, l := range logs {
		if l.TriggerType != "rabbitmq-subscribe" {
			t.Fatalf("unexpected triggerType: %s", l.TriggerType)
		}
		if l.Input["queue"] != "jobs" {
			t.Fatalf("expected queue in envelope, got %#v", l.Input)
		}
	}
	if len(ack.acked) != 2 {
		t.Fatalf("expected 2 acked deliveries, got %d", len(ack.acked))
	}
}

func TestBrokerSubscriber_StopCancelsSubscription(t *testing.T) {
	r := &fakeKafkaReader{queue: []kafka.Message{{Topic: "t1"}}}
	withFakeKafkaReader(t, r)

	engine := newTestIngestEngine(t)
	sub := NewBrokerSubscriber()
	sub.SetEngine(engine)

	if err := sub.StartKafkaSubscribe(context.Background(), "stop-me", KafkaConfig{Brokers: "b:9092", Topic: "t1"}, "ingest_test"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	waitForRuns(t, engine, 1)

	sub.Stop("stop-me")

	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) && !r.closed {
		time.Sleep(10 * time.Millisecond)
	}
	if !r.closed {
		t.Fatal("expected reader to be closed after Stop")
	}
}
