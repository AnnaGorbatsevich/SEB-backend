package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

type TelemetryEvent struct {
	Event        string          `json:"event"`
	Timestamp    string          `json:"timestamp"`
	SessionID    string          `json:"sessionId"`
	CustomUserID string          `json:"customUserId"`
	Data         json.RawMessage `json:"data"`
	ReceivedAt   time.Time       `json:"received_at"`
}

type CursorData struct {
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
	Ts string  `json:"ts"`
}

type KeyPressData struct {
	KeyCode   int      `json:"keyCode"`
	KeyName   string   `json:"keyName"`
	Modifiers []string `json:"modifiers"`
	IsCombo   bool     `json:"isCombo"`
	Ts        string   `json:"ts"`
}

type LogData struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Ts      string `json:"ts"`
}

type DiagnosticData struct {
	Code    string          `json:"code"`
	Status  string          `json:"status"`
	Details json.RawMessage `json:"details"`
}

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(broker, topic string) *Producer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	log.Printf("[KAFKA] producer initialised: broker=%s topic=%s", broker, topic)
	return &Producer{writer: w}
}

func (p *Producer) Publish(ctx context.Context, event TelemetryEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.Event),
		Value: payload,
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
