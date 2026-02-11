package messaging

import (
	"encoding/json/v2"
	"time"

	"github.com/google/uuid"
)

const CurrentVersion = "1.0"

// Envelope wraps any payload with metadata for tracing and versioning.
type Envelope[T any] struct {
	Version   string    `json:"version"`
	MessageID string    `json:"message_id"`
	Timestamp time.Time `json:"timestamp"`
	Payload   T         `json:"payload"`
}

// NewEnvelope creates an Envelope with a generated MessageID and current timestamp.
func NewEnvelope[T any](payload T) Envelope[T] {
	return Envelope[T]{
		Version:   CurrentVersion,
		MessageID: uuid.New().String(),
		Timestamp: time.Now().UTC(),
		Payload:   payload,
	}
}

// Marshal serialises an Envelope to JSON bytes.
func Marshal[T any](e Envelope[T]) ([]byte, error) {
	return json.Marshal(e)
}

// Unmarshal deserialises JSON bytes into an Envelope.
func Unmarshal[T any](data []byte) (Envelope[T], error) {
	var e Envelope[T]
	if err := json.Unmarshal(data, &e); err != nil {
		return e, err
	}
	return e, nil
}
