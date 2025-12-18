package pubsub

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestNewBroker verifies broker initialization
func TestNewBroker(t *testing.T) {
	broker := NewBroker[string]()

	if broker == nil {
		t.Fatal("NewBroker returned nil")
	}

	if broker.subs == nil {
		t.Error("subs map not initialized")
	}

	if broker.done == nil {
		t.Error("done channel not initialized")
	}

	if broker.subCount != 0 {
		t.Errorf("expected subCount 0, got %d", broker.subCount)
	}

	if broker.maxEvents != 1000 {
		t.Errorf("expected maxEvents 1000, got %d", broker.maxEvents)
	}
}

// TestNewBrokerWithOptions verifies custom initialization
func TestNewBrokerWithOptions(t *testing.T) {
	bufferSize := 128
	maxEvents := 5000

	broker := NewBrokerWithOptions[int](bufferSize, maxEvents)

	if broker == nil {
		t.Fatal("NewBrokerWithOptions returned nil")
	}

	if broker.maxEvents != maxEvents {
		t.Errorf("expected maxEvents %d, got %d", maxEvents, broker.maxEvents)
	}
}

// TestBrokerSubscribe verifies subscription functionality
func TestBrokerSubscribe(t *testing.T) {
	broker := NewBroker[string]()
	defer broker.Shutdown()

	ctx := context.Background()
	sub := broker.Subscribe(ctx)

	if sub == nil {
		t.Fatal("Subscribe returned nil channel")
	}

	count := broker.GetSubscriberCount()
	if count != 1 {
		t.Errorf("expected 1 subscriber, got %d", count)
	}
}

// TestBrokerMultipleSubscribers verifies multiple subscriptions
func TestBrokerMultipleSubscribers(t *testing.T) {
	broker := NewBroker[string]()
	defer broker.Shutdown()

	ctx := context.Background()

	sub1 := broker.Subscribe(ctx)
	sub2 := broker.Subscribe(ctx)
	sub3 := broker.Subscribe(ctx)

	if sub1 == nil || sub2 == nil || sub3 == nil {
		t.Fatal("Subscribe returned nil channel")
	}

	count := broker.GetSubscriberCount()
	if count != 3 {
		t.Errorf("expected 3 subscribers, got %d", count)
	}
}

// TestBrokerPublish verifies event publishing
func TestBrokerPublish(t *testing.T) {
	broker := NewBroker[string]()
	defer broker.Shutdown()

	ctx := context.Background()
	sub := broker.Subscribe(ctx)

	// Publish an event
	testPayload := "test message"
	broker.Publish(CreatedEvent, testPayload)

	// Receive the event
	select {
	case event := <-sub:
		if event.Type != CreatedEvent {
			t.Errorf("expected event type %s, got %s", CreatedEvent, event.Type)
		}
		if event.Payload != testPayload {
			t.Errorf("expected payload %s, got %s", testPayload, event.Payload)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

// TestBrokerPublishToMultipleSubscribers verifies broadcast functionality
func TestBrokerPublishToMultipleSubscribers(t *testing.T) {
	broker := NewBroker[int]()
	defer broker.Shutdown()

	ctx := context.Background()

	sub1 := broker.Subscribe(ctx)
	sub2 := broker.Subscribe(ctx)
	sub3 := broker.Subscribe(ctx)

	testPayload := 42
	broker.Publish(UpdatedEvent, testPayload)

	// All subscribers should receive the event
	var wg sync.WaitGroup
	wg.Add(3)

	checkEvent := func(sub <-chan Event[int], name string) {
		defer wg.Done()
		select {
		case event := <-sub:
			if event.Type != UpdatedEvent {
				t.Errorf("%s: expected event type %s, got %s", name, UpdatedEvent, event.Type)
			}
			if event.Payload != testPayload {
				t.Errorf("%s: expected payload %d, got %d", name, testPayload, event.Payload)
			}
		case <-time.After(1 * time.Second):
			t.Errorf("%s: timeout waiting for event", name)
		}
	}

	go checkEvent(sub1, "sub1")
	go checkEvent(sub2, "sub2")
	go checkEvent(sub3, "sub3")

	wg.Wait()
}

// TestBrokerShutdown verifies graceful shutdown
func TestBrokerShutdown(t *testing.T) {
	broker := NewBroker[string]()

	ctx := context.Background()
	sub := broker.Subscribe(ctx)

	broker.Shutdown()

	// Subscriber count should be 0
	count := broker.GetSubscriberCount()
	if count != 0 {
		t.Errorf("expected 0 subscribers after shutdown, got %d", count)
	}

	// Channel should be closed
	_, ok := <-sub
	if ok {
		t.Error("subscriber channel should be closed after shutdown")
	}
}

// TestBrokerShutdownIdempotent verifies shutdown can be called multiple times
func TestBrokerShutdownIdempotent(t *testing.T) {
	broker := NewBroker[string]()

	broker.Shutdown()
	broker.Shutdown() // Should not panic
	broker.Shutdown() // Should not panic
}

// TestBrokerSubscribeAfterShutdown verifies subscription after shutdown
func TestBrokerSubscribeAfterShutdown(t *testing.T) {
	broker := NewBroker[string]()
	broker.Shutdown()

	ctx := context.Background()
	sub := broker.Subscribe(ctx)

	// Should return closed channel
	_, ok := <-sub
	if ok {
		t.Error("subscription after shutdown should return closed channel")
	}
}

// TestBrokerPublishAfterShutdown verifies publish after shutdown doesn't panic
func TestBrokerPublishAfterShutdown(t *testing.T) {
	broker := NewBroker[string]()
	broker.Shutdown()

	// Should not panic
	broker.Publish(CreatedEvent, "test")
}

// TestBrokerContextCancellation verifies subscriber cleanup on context cancel
func TestBrokerContextCancellation(t *testing.T) {
	broker := NewBroker[string]()
	defer broker.Shutdown()

	ctx, cancel := context.WithCancel(context.Background())
	sub := broker.Subscribe(ctx)

	// Verify subscription
	count := broker.GetSubscriberCount()
	if count != 1 {
		t.Fatalf("expected 1 subscriber, got %d", count)
	}

	// Cancel context
	cancel()

	// Wait for cleanup (goroutine needs time to process cancellation)
	time.Sleep(50 * time.Millisecond)

	// Subscriber should be removed
	count = broker.GetSubscriberCount()
	if count != 0 {
		t.Errorf("expected 0 subscribers after context cancel, got %d", count)
	}

	// Channel should be closed
	_, ok := <-sub
	if ok {
		t.Error("subscriber channel should be closed after context cancel")
	}
}

// TestBrokerEventTypes verifies different event types
func TestBrokerEventTypes(t *testing.T) {
	broker := NewBroker[string]()
	defer broker.Shutdown()

	ctx := context.Background()
	sub := broker.Subscribe(ctx)

	tests := []struct {
		eventType EventType
		payload   string
	}{
		{CreatedEvent, "created payload"},
		{UpdatedEvent, "updated payload"},
		{DeletedEvent, "deleted payload"},
		{EventType("custom"), "custom payload"},
	}

	for _, tt := range tests {
		broker.Publish(tt.eventType, tt.payload)

		select {
		case event := <-sub:
			if event.Type != tt.eventType {
				t.Errorf("expected event type %s, got %s", tt.eventType, event.Type)
			}
			if event.Payload != tt.payload {
				t.Errorf("expected payload %s, got %s", tt.payload, event.Payload)
			}
		case <-time.After(1 * time.Second):
			t.Fatalf("timeout waiting for event type %s", tt.eventType)
		}
	}
}

// TestBrokerSlowSubscriber verifies handling of slow subscribers
func TestBrokerSlowSubscriber(t *testing.T) {
	broker := NewBrokerWithOptions[int](2, 1000) // Small buffer
	defer broker.Shutdown()

	ctx := context.Background()
	slowSub := broker.Subscribe(ctx)
	fastSub := broker.Subscribe(ctx)

	// Fill slow subscriber's buffer
	for i := 0; i < 5; i++ {
		broker.Publish(CreatedEvent, i)
	}

	// Fast subscriber drains immediately
	received := 0
	timeout := time.After(500 * time.Millisecond)

drainLoop:
	for {
		select {
		case <-fastSub:
			received++
		case <-timeout:
			break drainLoop
		default:
			// No more events immediately available
			break drainLoop
		}
	}

	// Fast subscriber should receive at least some events
	if received == 0 {
		t.Error("fast subscriber received no events")
	}

	// Slow subscriber buffer should have some events (but may drop some)
	slowReceived := 0
slowLoop:
	for {
		select {
		case <-slowSub:
			slowReceived++
		case <-time.After(100 * time.Millisecond):
			break slowLoop
		}
	}

	if slowReceived == 0 {
		t.Error("slow subscriber received no events")
	}
}

// TestBrokerConcurrentPublish verifies thread-safe publishing
func TestBrokerConcurrentPublish(t *testing.T) {
	broker := NewBroker[int]()
	defer broker.Shutdown()

	ctx := context.Background()
	sub := broker.Subscribe(ctx)

	const numPublishers = 5
	const numEventsPerPublisher = 50

	var wg sync.WaitGroup
	wg.Add(numPublishers)

	// Collect events in background (drain fast to avoid buffer overflow)
	received := make(map[int]bool)
	var receiveMu sync.Mutex
	done := make(chan struct{})

	go func() {
		for event := range sub {
			receiveMu.Lock()
			received[event.Payload] = true
			count := len(received)
			receiveMu.Unlock()

			if count >= numPublishers*numEventsPerPublisher {
				close(done)
				return
			}
		}
	}()

	// Start concurrent publishers after receiver is ready
	for i := 0; i < numPublishers; i++ {
		go func(publisherID int) {
			defer wg.Done()
			for j := 0; j < numEventsPerPublisher; j++ {
				broker.Publish(CreatedEvent, publisherID*1000+j)
				time.Sleep(time.Microsecond) // Small delay to avoid overwhelming buffer
			}
		}(i)
	}

	wg.Wait()

	// Give receiver time to process remaining events
	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		receiveMu.Lock()
		count := len(received)
		receiveMu.Unlock()

		// Some events may be dropped if buffer is full, that's expected behavior
		if count < numPublishers*numEventsPerPublisher/2 {
			t.Errorf("received too few events: %d (expected at least %d)",
				count, numPublishers*numEventsPerPublisher/2)
		}
	}
}

// TestBrokerGenericTypes verifies different payload types
func TestBrokerGenericTypes(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		broker := NewBroker[string]()
		defer broker.Shutdown()

		ctx := context.Background()
		sub := broker.Subscribe(ctx)

		broker.Publish(CreatedEvent, "test string")

		select {
		case event := <-sub:
			if event.Payload != "test string" {
				t.Errorf("expected 'test string', got %s", event.Payload)
			}
		case <-time.After(1 * time.Second):
			t.Fatal("timeout")
		}
	})

	t.Run("struct", func(t *testing.T) {
		type TestStruct struct {
			ID   int
			Name string
		}

		broker := NewBroker[TestStruct]()
		defer broker.Shutdown()

		ctx := context.Background()
		sub := broker.Subscribe(ctx)

		payload := TestStruct{ID: 42, Name: "test"}
		broker.Publish(CreatedEvent, payload)

		select {
		case event := <-sub:
			if event.Payload.ID != 42 || event.Payload.Name != "test" {
				t.Errorf("unexpected payload: %+v", event.Payload)
			}
		case <-time.After(1 * time.Second):
			t.Fatal("timeout")
		}
	})

	t.Run("pointer", func(t *testing.T) {
		broker := NewBroker[*int]()
		defer broker.Shutdown()

		ctx := context.Background()
		sub := broker.Subscribe(ctx)

		value := 42
		broker.Publish(CreatedEvent, &value)

		select {
		case event := <-sub:
			if event.Payload == nil || *event.Payload != 42 {
				t.Error("unexpected pointer payload")
			}
		case <-time.After(1 * time.Second):
			t.Fatal("timeout")
		}
	})
}

// TestUpdateAvailableMsg verifies the update message structure
func TestUpdateAvailableMsg(t *testing.T) {
	msg := UpdateAvailableMsg{
		CurrentVersion: "0.1.0",
		LatestVersion:  "0.2.0",
		IsDevelopment:  false,
	}

	if msg.CurrentVersion != "0.1.0" {
		t.Errorf("expected CurrentVersion '0.1.0', got %s", msg.CurrentVersion)
	}

	if msg.LatestVersion != "0.2.0" {
		t.Errorf("expected LatestVersion '0.2.0', got %s", msg.LatestVersion)
	}

	if msg.IsDevelopment {
		t.Error("expected IsDevelopment false")
	}
}
