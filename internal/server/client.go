package server

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

var tmpl *template.Template

const (
	WriteWait  = 10 * time.Second
	PongWait   = 60 * time.Second
	PingPeriod = (PongWait * 9) / 10
	MaxSize    = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

func newMessageID() (string, error) {
	b := make([]byte, 16)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func (c *Client) ReadPump(msg chan<- Message) {
	defer func() {
		c.Hub.unregister <- c
		if err := c.conn.Close(); err != nil {
			log.Printf("read close: %v", err)
		}
	}()

	c.conn.SetReadLimit(MaxSize)

	if err := c.conn.SetReadDeadline(time.Now().Add(PongWait)); err != nil {
		log.Printf("SetReadDeadline: %v", err)
		return
	}

	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(PongWait))
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket read: %v", err)
			}
			break
		}

		message = bytes.TrimSpace(message)

		var event ClientEvent
		if err := json.Unmarshal(message, &event); err != nil {
			log.Printf("failed to decode client event: %v", err)
			continue
		}

		switch event.Type {
		case EventMessage:
			if event.Message == "" {
				continue
			}

			messageID, err := newMessageID()
			if err != nil {
				log.Printf("failed to generate message ID: %v", err)
				continue
			}

			formattedTime := time.Now().Format("2-Jan-06 15:04:05")

			form := Message{
				MessageID:  messageID,
				RoomCode:   c.roomcode,
				Username:   c.username,
				Message:    event.Message,
				Time_Stamp: formattedTime,
			}

			var buf bytes.Buffer
			if err := tmpl.ExecuteTemplate(&buf, "message", form); err != nil {
				log.Printf("failed to render message: %v", err)
				continue
			}

			msg <- form

			c.Hub.broadcast <- Event{
				Type:      "text",
				MessageID: messageID,
				RoomCode:  c.roomcode,
				Username:  c.username,
				Content:   form.Message,
				TimeStamp: formattedTime,
				HTML:      buf.String(),
				Sender:    c,
			}

		case EventDelivered, EventRead:
			if event.MessageID == "" {
				continue
			}

			c.Hub.broadcast <- Event{
				Type:      event.Type,
				MessageID: event.MessageID,
				RoomCode:  c.roomcode,
				Username:  c.username,
			}

		default:
			log.Printf("unknown event type: %q", event.Type)
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(PingPeriod)

	defer func() {
		ticker.Stop()
		if err := c.conn.Close(); err != nil {
			log.Printf("write close: %v", err)
		}
	}()

	for {
		select {
		case event, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(WriteWait)); err != nil {
				log.Printf("SetWriteDeadline: %v", err)
				return
			}

			if !ok {
				if err := c.conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Printf("close message: %v", err)
				}
				return
			}

			var payload []byte

			if event.Type == "text" {
				payload = []byte(event.HTML)
			} else {
				encoded, err := json.Marshal(event)
				if err != nil {
					log.Printf("failed to encode event: %v", err)
					continue
				}
				payload = encoded
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				log.Printf("failed to write event: %v", err)
				return
			}

		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(WriteWait)); err != nil {
				log.Printf("SetWriteDeadline: %v", err)
				return
			}

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("ping: %v", err)
				return
			}
		}
	}
}
