package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	kafka "github.com/segmentio/kafka-go"
	"studentstorage/internal/db"
)

type TelemetryEvent struct {
	Event     string          `json:"event"`
	Timestamp string          `json:"timestamp"`
	SessionID string          `json:"sessionId"`
	Email     string          `json:"customUserId"`
	Data      json.RawMessage `json:"data"`
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

func marshalModifiers(m []string) string {
	b, _ := json.Marshal(m)
	return string(b)
}

func unmarshalModifiers(s string) []string {
	var m []string
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return []string{}
	}
	return m
}

func Consume(ctx context.Context, broker, topic, groupID string, dbService *db.Service) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:           []string{broker},
		Topic:             topic,
		GroupID:           groupID,
		MinBytes:          1,
		MaxBytes:          10e6,
		HeartbeatInterval: 3 * time.Second,
		SessionTimeout:    30 * time.Second,
		RebalanceTimeout:  30 * time.Second,
		StartOffset:       kafka.FirstOffset,
	})
	defer r.Close()

	log.Printf("[KAFKA] consumer started: topic=%s group=%s", topic, groupID)

	for {
		msg, err := r.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("[KAFKA] read error: %v", err)
			continue
		}

		var event TelemetryEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("[KAFKA] failed to parse message: %v", err)
			continue
		}

		switch event.Event {
		case "cursor_position":
			storeCursor(ctx, event, dbService)
		case "key_press":
			storeKeyPress(ctx, event, dbService)
		case "log":
			storeLog(ctx, event, dbService)
		default:
			log.Printf("[KAFKA] unknown event type: %s", event.Event)
		}
	}
}

func storeCursor(ctx context.Context, event TelemetryEvent, dbService *db.Service) {
	var cursor CursorData
	if err := json.Unmarshal(event.Data, &cursor); err != nil {
		log.Printf("[KAFKA] failed to parse cursor data: %v", err)
		return
	}

	ts, err := time.Parse(time.RFC3339, cursor.Ts)
	if err != nil {
		ts, _ = time.Parse(time.RFC3339, event.Timestamp)
	}

	_, err = dbService.DB.ExecContext(ctx, `
		INSERT INTO cursor_positions (session_id, x, y, ts, email)
		VALUES ($1, $2, $3, $4, $5)
	`, event.SessionID, cursor.X, cursor.Y, ts, event.Email)
	if err != nil {
		log.Printf("[DB] insert cursor error: %v", err)
		return
	}

	log.Printf("[STORE] cursor session=%s x=%.1f y=%.1f ts=%s", event.SessionID, cursor.X, cursor.Y, ts.Format(time.RFC3339))
}

func storeKeyPress(ctx context.Context, event TelemetryEvent, dbService *db.Service) {
	var kp KeyPressData
	if err := json.Unmarshal(event.Data, &kp); err != nil {
		log.Printf("[KAFKA] failed to parse key_press data: %v", err)
		return
	}

	ts, err := time.Parse(time.RFC3339, kp.Ts)
	if err != nil {
		ts, _ = time.Parse(time.RFC3339, event.Timestamp)
	}

	if kp.Modifiers == nil {
		kp.Modifiers = []string{}
	}

	_, err = dbService.DB.ExecContext(ctx, `
		INSERT INTO key_presses (session_id, key_code, key_name, modifiers, is_combo, ts, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, event.SessionID, kp.KeyCode, kp.KeyName, marshalModifiers(kp.Modifiers), kp.IsCombo, ts, event.Email)
	if err != nil {
		log.Printf("[DB] insert key_press error: %v", err)
		return
	}

	log.Printf("[STORE] key_press session=%s key=%s(%d) modifiers=%v isCombo=%v ts=%s",
		event.SessionID, kp.KeyName, kp.KeyCode, kp.Modifiers, kp.IsCombo, ts.Format(time.RFC3339))
}

func storeLog(ctx context.Context, event TelemetryEvent, dbService *db.Service) {
	var l LogData
	if err := json.Unmarshal(event.Data, &l); err != nil {
		log.Printf("[KAFKA] failed to parse log data: %v", err)
		return
	}

	ts, err := time.Parse(time.RFC3339, l.Ts)
	if err != nil {
		ts, _ = time.Parse(time.RFC3339, event.Timestamp)
	}

	_, err = dbService.DB.ExecContext(ctx, `
		INSERT INTO logs (session_id, level, message, ts, email)
		VALUES ($1, $2, $3, $4, $5)
	`, event.SessionID, l.Level, l.Message, ts, event.Email)
	if err != nil {
		log.Printf("[DB] insert log error: %v", err)
		return
	}

	log.Printf("[STORE] log session=%s level=%s message=\"%s\" ts=%s", event.SessionID, l.Level, l.Message, ts.Format(time.RFC3339))
}