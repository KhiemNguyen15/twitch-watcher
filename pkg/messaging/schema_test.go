package messaging_test

import (
	"testing"
	"time"

	"github.com/khiemnguyen15/twitch-watcher/pkg/messaging"
)

type testPayload struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestNewEnvelope(t *testing.T) {
	before := time.Now().UTC()
	payload := testPayload{Name: "hello", Value: 42}
	env := messaging.NewEnvelope(payload)
	after := time.Now().UTC()

	if env.Version != messaging.CurrentVersion {
		t.Errorf("Version = %q, want %q", env.Version, messaging.CurrentVersion)
	}
	if env.MessageID == "" {
		t.Error("MessageID must not be empty")
	}
	if env.Timestamp.Before(before) || env.Timestamp.After(after) {
		t.Errorf("Timestamp %v not in expected range [%v, %v]", env.Timestamp, before, after)
	}
	if env.Payload != payload {
		t.Errorf("Payload = %+v, want %+v", env.Payload, payload)
	}
}

func TestNewEnvelope_UniqueMessageIDs(t *testing.T) {
	e1 := messaging.NewEnvelope(testPayload{})
	e2 := messaging.NewEnvelope(testPayload{})
	if e1.MessageID == e2.MessageID {
		t.Error("consecutive envelopes must have different MessageIDs")
	}
}

func TestMarshalUnmarshal_RoundTrip(t *testing.T) {
	original := messaging.NewEnvelope(testPayload{Name: "round-trip", Value: 99})

	data, err := messaging.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	decoded, err := messaging.Unmarshal[testPayload](data)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.Version != original.Version {
		t.Errorf("Version: got %q, want %q", decoded.Version, original.Version)
	}
	if decoded.MessageID != original.MessageID {
		t.Errorf("MessageID: got %q, want %q", decoded.MessageID, original.MessageID)
	}
	if !decoded.Timestamp.Equal(original.Timestamp) {
		t.Errorf("Timestamp: got %v, want %v", decoded.Timestamp, original.Timestamp)
	}
	if decoded.Payload != original.Payload {
		t.Errorf("Payload: got %+v, want %+v", decoded.Payload, original.Payload)
	}
}

func TestUnmarshal_InvalidJSON(t *testing.T) {
	_, err := messaging.Unmarshal[testPayload]([]byte(`{invalid}`))
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}
