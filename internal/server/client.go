package server

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func (c *Client) ReadPump(msg chan<- Message) {
	defer func() {
		c.Hub.unregister <- c
		err := c.conn.Close()
		if err != nil {
			log.Printf("read close %v", err)
		}
	}()

	c.conn.SetReadLimit(MaxSize)
	err := c.conn.SetReadDeadline(time.Now().Add(PongWait))
	if err != nil {
		log.Printf("SetReadDeadline error %v", err)
	}
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(PongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			fmt.Printf("%v", err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("%v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		var form Message

		err = json.Unmarshal(message, &form)
		if err != nil {
			fmt.Printf("failed to decode message: %v\n", err)
			continue
		}

		origTime := time.Now()
		layout := "2-Jan-06 15:04:05"
		formattedTime := origTime.Format(layout)

		form.Username = c.username
		form.Time_Stamp = formattedTime
		var buf bytes.Buffer

		err = tmpl.ExecuteTemplate(&buf, "message", form)
		if err != nil {
			fmt.Printf("failed to render room.tmpl: %v\n", err)
			continue
		}

		msg <- Message{
			RoomCode:   c.roomcode,
			Username:   c.username,
			Message:    form.Message,
			Time_Stamp: formattedTime,
		}

		event := Event{
			Type:      "text",
			RoomCode:  c.roomcode,
			Username:  c.username,
			Content:   form.Message,
			TimeStamp: form.Time_Stamp,
			HTML:      buf.String(),
		}

		c.Hub.broadcast <- event

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

			if err := c.conn.SetWriteDeadline(
				time.Now().Add(WriteWait),
			); err != nil {
				log.Printf("SetWriteDeadline: %v", err)
				return
			}

			if !ok {
				if err := c.conn.WriteMessage(
					websocket.CloseMessage,
					[]byte{},
				); err != nil {
					log.Printf("close message: %v", err)
				}
				return
			}

			if err := c.conn.WriteMessage(
				websocket.TextMessage,
				[]byte(event.HTML),
			); err != nil {
				log.Printf("failed to write event: %v", err)
				return
			}

		case <-ticker.C:

			if err := c.conn.SetWriteDeadline(
				time.Now().Add(WriteWait),
			); err != nil {
				log.Printf("SetWriteDeadline: %v", err)
				return
			}

			if err := c.conn.WriteMessage(
				websocket.PingMessage,
				nil,
			); err != nil {
				log.Printf("ping: %v", err)
				return
			}
		}
	}
}
