package rulesgo

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

// --- fakes: no live Kafka cluster required ---

type fakeKafkaWriter struct {
	mu     sync.Mutex
	msgs   []kafka.Message
	closed bool
}

func (f *fakeKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.msgs = append(f.msgs, msgs...)
	return nil
}
func (f *fakeKafkaWriter) Close() error { f.closed = true; return nil }

type fakeKafkaReader struct {
	mu        sync.Mutex
	queue     []kafka.Message
	committed []kafka.Message
	closed    bool
}

func (f *fakeKafkaReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	f.mu.Lock()
	if len(f.queue) == 0 {
		f.mu.Unlock()
		// simulate "no more messages waiting right now" - block until the
		// caller's timeout/cancellation fires, same as a real kafka.Reader.
		<-ctx.Done()
		return kafka.Message{}, ctx.Err()
	}
	m := f.queue[0]
	f.queue = f.queue[1:]
	f.mu.Unlock()
	return m, nil
}
func (f *fakeKafkaReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.committed = append(f.committed, msgs...)
	return nil
}
func (f *fakeKafkaReader) Close() error { f.closed = true; return nil }

func withFakeKafkaWriter(t *testing.T, w *fakeKafkaWriter) {
	t.Helper()
	orig := newKafkaWriter
	newKafkaWriter = func(cfg KafkaConfig) kafkaWriter { return w }
	t.Cleanup(func() { newKafkaWriter = orig })
}

func withFakeKafkaReader(t *testing.T, r *fakeKafkaReader) {
	t.Helper()
	orig := newKafkaReader
	newKafkaReader = func(cfg KafkaConfig) kafkaReader { return r }
	t.Cleanup(func() { newKafkaReader = orig })
}

func TestKafkaProduce_WritesResolvedMessage(t *testing.T) {
	w := &fakeKafkaWriter{}
	withFakeKafkaWriter(t, w)

	n := &kafkaProduceExecutor{}
	node := ChainNode{ID: "n1", Type: "kafka.produce", Config: json.RawMessage(`{
		"brokers": "broker1:9092,broker2:9092",
		"topic": "orders",
		"key": "{{orderID}}",
		"value": "{{payload}}"
	}`)}
	ec := &ExecutionContext{Variables: map[string]interface{}{"orderID": "42", "payload": `{"amount":10}`}}

	out, err := n.Execute(context.Background(), node, ec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["success"] != true {
		t.Fatalf("expected success, got %#v", out)
	}
	if len(w.msgs) != 1 {
		t.Fatalf("expected 1 message written, got %d", len(w.msgs))
	}
	if string(w.msgs[0].Key) != "42" {
		t.Fatalf("key not resolved: %q", w.msgs[0].Key)
	}
	if string(w.msgs[0].Value) != `{"amount":10}` {
		t.Fatalf("value not resolved: %q", w.msgs[0].Value)
	}
	if !w.closed {
		t.Fatal("writer was not closed")
	}
}

func TestKafkaProduce_RequiresBrokersAndTopic(t *testing.T) {
	withFakeKafkaWriter(t, &fakeKafkaWriter{})
	n := &kafkaProduceExecutor{}

	if _, err := n.Execute(context.Background(), ChainNode{Config: json.RawMessage(`{"topic":"x","value":"v"}`)}, &ExecutionContext{}); err == nil {
		t.Fatal("expected error for missing brokers")
	}
	if _, err := n.Execute(context.Background(), ChainNode{Config: json.RawMessage(`{"brokers":"b:9092","value":"v"}`)}, &ExecutionContext{}); err == nil {
		t.Fatal("expected error for missing topic")
	}
}

func TestKafkaConsume_BatchesUpToMaxMessagesAndCommits(t *testing.T) {
	r := &fakeKafkaReader{queue: []kafka.Message{
		{Key: []byte("k1"), Value: []byte("v1"), Partition: 0, Offset: 1},
		{Key: []byte("k2"), Value: []byte("v2"), Partition: 0, Offset: 2},
		{Key: []byte("k3"), Value: []byte("v3"), Partition: 0, Offset: 3},
	}}
	withFakeKafkaReader(t, r)

	n := &kafkaConsumeExecutor{}
	node := ChainNode{Config: json.RawMessage(`{"brokers":"b:9092","topic":"orders","maxMessages":2,"timeoutSeconds":1}`)}

	out, err := n.Execute(context.Background(), node, &ExecutionContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["count"] != 2 {
		t.Fatalf("expected 2 messages, got %#v", out["count"])
	}
	msgs, _ := out["messages"].([]map[string]interface{})
	if len(msgs) != 2 || msgs[0]["key"] != "k1" || msgs[1]["key"] != "k2" {
		t.Fatalf("unexpected messages: %#v", msgs)
	}
	if len(r.committed) != 2 {
		t.Fatalf("expected 2 committed messages, got %d", len(r.committed))
	}
	if !r.closed {
		t.Fatal("reader was not closed")
	}
}

func TestKafkaConsume_StopsOnTimeoutWithFewerThanMax(t *testing.T) {
	r := &fakeKafkaReader{queue: []kafka.Message{{Key: []byte("only")}}}
	withFakeKafkaReader(t, r)

	n := &kafkaConsumeExecutor{}
	node := ChainNode{Config: json.RawMessage(`{"brokers":"b:9092","topic":"orders","maxMessages":5,"timeoutSeconds":1}`)}

	start := time.Now()
	out, err := n.Execute(context.Background(), node, &ExecutionContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["count"] != 1 {
		t.Fatalf("expected 1 message, got %#v", out["count"])
	}
	if time.Since(start) > 3*time.Second {
		t.Fatalf("took too long: %s", time.Since(start))
	}
}

func TestKafkaSubscribe_CallsStartWithResolvedConfig(t *testing.T) {
	var gotSubKey string
	var gotCfg KafkaConfig
	var gotIngest string
	n := &kafkaSubscribeExecutor{start: func(ctx context.Context, subKey string, cfg KafkaConfig, ingestChainID string) error {
		gotSubKey, gotCfg, gotIngest = subKey, cfg, ingestChainID
		return nil
	}}

	node := ChainNode{ID: "sub1", Config: json.RawMessage(`{
		"brokers": "b:9092",
		"topic": "{{topic}}",
		"ingestChainID": "ingest_orders"
	}`)}
	ec := &ExecutionContext{Variables: map[string]interface{}{"topic": "orders"}}

	out, err := n.Execute(context.Background(), node, ec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["subscribed"] != true {
		t.Fatalf("expected subscribed=true, got %#v", out)
	}
	if gotCfg.Topic != "orders" {
		t.Fatalf("topic not resolved: %q", gotCfg.Topic)
	}
	if gotIngest != "ingest_orders" {
		t.Fatalf("unexpected ingestChainID: %q", gotIngest)
	}
	if gotSubKey != "sub1:orders" {
		t.Fatalf("unexpected subKey: %q", gotSubKey)
	}
}

func TestKafkaSubscribe_NotConfigured(t *testing.T) {
	n := &kafkaSubscribeExecutor{}
	node := ChainNode{ID: "sub1", Config: json.RawMessage(`{"brokers":"b:9092","topic":"orders","ingestChainID":"x"}`)}
	out, err := n.Execute(context.Background(), node, &ExecutionContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["status"] != "kafka_subscribe_not_configured" {
		t.Fatalf("unexpected output: %#v", out)
	}
}

func TestKafkaSubscribe_RequiresIngestChainID(t *testing.T) {
	n := &kafkaSubscribeExecutor{start: func(ctx context.Context, subKey string, cfg KafkaConfig, ingestChainID string) error {
		t.Fatal("start should not be called without ingestChainID")
		return nil
	}}
	node := ChainNode{ID: "sub1", Config: json.RawMessage(`{"brokers":"b:9092","topic":"orders"}`)}
	if _, err := n.Execute(context.Background(), node, &ExecutionContext{}); err == nil {
		t.Fatal("expected error for missing ingestChainID")
	}
}
