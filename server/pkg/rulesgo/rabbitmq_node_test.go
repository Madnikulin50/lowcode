package rulesgo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/madnikulin50/lowcode/server/pkg/vault"
)

// --- fakes: no live RabbitMQ server required ---

type fakeAcknowledger struct {
	mu     sync.Mutex
	acked  []uint64
	nacked []uint64
}

func (a *fakeAcknowledger) Ack(tag uint64, multiple bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.acked = append(a.acked, tag)
	return nil
}
func (a *fakeAcknowledger) Nack(tag uint64, multiple bool, requeue bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.nacked = append(a.nacked, tag)
	return nil
}
func (a *fakeAcknowledger) Reject(tag uint64, requeue bool) error { return nil }

type fakeRabbitMQChannel struct {
	mu         sync.Mutex
	published  []amqp.Publishing
	exchange   string
	routingKey string
	deliveries chan amqp.Delivery
	consumeErr error
	closed     bool
	publishErr error
}

func (f *fakeRabbitMQChannel) PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	if f.publishErr != nil {
		return f.publishErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.exchange = exchange
	f.routingKey = key
	f.published = append(f.published, msg)
	return nil
}

func (f *fakeRabbitMQChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	if f.consumeErr != nil {
		return nil, f.consumeErr
	}
	return f.deliveries, nil
}

func (f *fakeRabbitMQChannel) Close() error { f.closed = true; return nil }

type fakeRabbitMQConnection struct {
	ch         *fakeRabbitMQChannel
	channelErr error
	closed     bool
}

func (f *fakeRabbitMQConnection) Channel() (rabbitmqChannel, error) {
	if f.channelErr != nil {
		return nil, f.channelErr
	}
	return f.ch, nil
}
func (f *fakeRabbitMQConnection) Close() error { f.closed = true; return nil }

func withFakeRabbitMQDialer(t *testing.T, conn *fakeRabbitMQConnection) {
	t.Helper()
	orig := dialRabbitMQ
	dialRabbitMQ = func(url string) (rabbitmqConnection, error) { return conn, nil }
	t.Cleanup(func() { dialRabbitMQ = orig })
}

func TestRabbitMQPublish_PublishesResolvedMessage(t *testing.T) {
	ch := &fakeRabbitMQChannel{}
	conn := &fakeRabbitMQConnection{ch: ch}
	withFakeRabbitMQDialer(t, conn)

	n := &rabbitmqPublishExecutor{}
	node := ChainNode{Config: json.RawMessage(`{
		"url": "amqp://guest:guest@localhost:5672/",
		"exchange": "orders.exchange",
		"routingKey": "{{region}}",
		"body": "{{payload}}"
	}`)}
	ec := &ExecutionContext{Variables: map[string]interface{}{"region": "eu", "payload": `{"id":1}`}}

	out, err := n.Execute(context.Background(), node, ec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["success"] != true {
		t.Fatalf("expected success, got %#v", out)
	}
	if len(ch.published) != 1 {
		t.Fatalf("expected 1 published message, got %d", len(ch.published))
	}
	if ch.exchange != "orders.exchange" || ch.routingKey != "eu" {
		t.Fatalf("unexpected exchange/routingKey: %q/%q", ch.exchange, ch.routingKey)
	}
	if string(ch.published[0].Body) != `{"id":1}` {
		t.Fatalf("body not resolved: %q", ch.published[0].Body)
	}
	if !ch.closed || !conn.closed {
		t.Fatal("channel/connection were not closed")
	}
}

func TestRabbitMQPublish_FallsBackToQueueAsRoutingKey(t *testing.T) {
	ch := &fakeRabbitMQChannel{}
	conn := &fakeRabbitMQConnection{ch: ch}
	withFakeRabbitMQDialer(t, conn)

	n := &rabbitmqPublishExecutor{}
	node := ChainNode{Config: json.RawMessage(`{"url":"amqp://x","queue":"jobs","body":"hi"}`)}

	if _, err := n.Execute(context.Background(), node, &ExecutionContext{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ch.routingKey != "jobs" {
		t.Fatalf("expected routingKey to fall back to queue, got %q", ch.routingKey)
	}
}

func TestRabbitMQPublish_URLSecretRefOverridesPlaintextURL(t *testing.T) {
	vaultSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"data": map[string]interface{}{"url": "amqp://real:secret@broker/"},
			},
		})
	}))
	defer vaultSrv.Close()
	restore := vault.SetDefault(&vault.Client{Addr: vaultSrv.URL, Token: "t", Mount: "secret"})
	defer restore()

	ch := &fakeRabbitMQChannel{}
	var dialedURL string
	origDial := dialRabbitMQ
	dialRabbitMQ = func(url string) (rabbitmqConnection, error) {
		dialedURL = url
		return &fakeRabbitMQConnection{ch: ch}, nil
	}
	defer func() { dialRabbitMQ = origDial }()

	n := &rabbitmqPublishExecutor{}
	node := ChainNode{Config: json.RawMessage(`{
		"url": "amqp://placeholder-should-be-overridden",
		"urlSecretRef": "compose/connectors/1#url",
		"queue": "jobs",
		"body": "hi"
	}`)}

	if _, err := n.Execute(context.Background(), node, &ExecutionContext{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dialedURL != "amqp://real:secret@broker/" {
		t.Fatalf("expected dial to use the vault-resolved URL, got %q", dialedURL)
	}
}

func TestRabbitMQPublish_RequiresURLAndRoutingTarget(t *testing.T) {
	ch := &fakeRabbitMQChannel{}
	conn := &fakeRabbitMQConnection{ch: ch}
	withFakeRabbitMQDialer(t, conn)
	n := &rabbitmqPublishExecutor{}

	if _, err := n.Execute(context.Background(), ChainNode{Config: json.RawMessage(`{"queue":"q","body":"x"}`)}, &ExecutionContext{}); err == nil {
		t.Fatal("expected error for missing url")
	}
	if _, err := n.Execute(context.Background(), ChainNode{Config: json.RawMessage(`{"url":"amqp://x","body":"x"}`)}, &ExecutionContext{}); err == nil {
		t.Fatal("expected error for missing routingKey/queue")
	}
}

func TestRabbitMQConsume_AcksAndReturnsMessages(t *testing.T) {
	ack := &fakeAcknowledger{}
	deliveries := make(chan amqp.Delivery, 2)
	deliveries <- amqp.Delivery{Body: []byte("m1"), RoutingKey: "rk", DeliveryTag: 1, Acknowledger: ack}
	deliveries <- amqp.Delivery{Body: []byte("m2"), RoutingKey: "rk", DeliveryTag: 2, Acknowledger: ack}
	close(deliveries) // simulate the consumer channel ending after 2 messages

	ch := &fakeRabbitMQChannel{deliveries: deliveries}
	conn := &fakeRabbitMQConnection{ch: ch}
	withFakeRabbitMQDialer(t, conn)

	n := &rabbitmqConsumeExecutor{}
	node := ChainNode{Config: json.RawMessage(`{"url":"amqp://x","queue":"jobs","maxMessages":10,"timeoutSeconds":2}`)}

	out, err := n.Execute(context.Background(), node, &ExecutionContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["count"] != 2 {
		t.Fatalf("expected 2 messages, got %#v", out["count"])
	}
	if len(ack.acked) != 2 {
		t.Fatalf("expected 2 acked deliveries, got %d", len(ack.acked))
	}
}

func TestRabbitMQConsume_StopsAtTimeoutBelowMax(t *testing.T) {
	ack := &fakeAcknowledger{}
	deliveries := make(chan amqp.Delivery, 1)
	deliveries <- amqp.Delivery{Body: []byte("only"), DeliveryTag: 1, Acknowledger: ack}
	// do NOT close - simulate a live queue with just one message waiting

	ch := &fakeRabbitMQChannel{deliveries: deliveries}
	conn := &fakeRabbitMQConnection{ch: ch}
	withFakeRabbitMQDialer(t, conn)

	n := &rabbitmqConsumeExecutor{}
	node := ChainNode{Config: json.RawMessage(`{"url":"amqp://x","queue":"jobs","maxMessages":5,"timeoutSeconds":1}`)}

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

func TestRabbitMQSubscribe_CallsStartWithResolvedConfig(t *testing.T) {
	var gotSubKey, gotIngest string
	var gotCfg RabbitMQConfig
	n := &rabbitmqSubscribeExecutor{start: func(ctx context.Context, subKey string, cfg RabbitMQConfig, ingestChainID string) error {
		gotSubKey, gotCfg, gotIngest = subKey, cfg, ingestChainID
		return nil
	}}

	node := ChainNode{ID: "sub1", Config: json.RawMessage(`{
		"url": "amqp://x",
		"queue": "{{queue}}",
		"ingestChainID": "ingest_jobs"
	}`)}
	ec := &ExecutionContext{Variables: map[string]interface{}{"queue": "jobs"}}

	out, err := n.Execute(context.Background(), node, ec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["subscribed"] != true {
		t.Fatalf("expected subscribed=true, got %#v", out)
	}
	if gotCfg.Queue != "jobs" || gotIngest != "ingest_jobs" || gotSubKey != "sub1:jobs" {
		t.Fatalf("unexpected resolved values: cfg=%#v ingest=%q subKey=%q", gotCfg, gotIngest, gotSubKey)
	}
}

func TestRabbitMQSubscribe_NotConfigured(t *testing.T) {
	n := &rabbitmqSubscribeExecutor{}
	node := ChainNode{ID: "sub1", Config: json.RawMessage(`{"url":"amqp://x","queue":"jobs","ingestChainID":"x"}`)}
	out, err := n.Execute(context.Background(), node, &ExecutionContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["status"] != "rabbitmq_subscribe_not_configured" {
		t.Fatalf("unexpected output: %#v", out)
	}
}
