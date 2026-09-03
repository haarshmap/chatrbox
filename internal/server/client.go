package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
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

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(MaxSize)
	c.conn.SetReadDeadline(time.Now().Add(PongWait))
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

		form.Username = c.username
		var buf bytes.Buffer

		err = tmpl.ExecuteTemplate(&buf, "message", form)
		if err != nil {
			fmt.Printf("failed to render room.tmpl: %v\n", err)
			continue
		}

		c.Hub.broadcast <- buf.Bytes()
	}
}

func (c *Client) WritePump() {
	Tick := time.NewTicker(PingPeriod)
	defer func() {
		Tick.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case Message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				fmt.Printf("%v", err)
				return
			}
			w.Write(Message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				fmt.Printf("%v", err)
				return
			}
		case <-Tick.C:
			c.conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				fmt.Printf("%v", err)
				return
			}
		}
	}
}
