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
			log.Fatalf("read close %v", err)
		}
	}()

	c.conn.SetReadLimit(MaxSize)
	err := c.conn.SetReadDeadline(time.Now().Add(PongWait))
	if err != nil {
		log.Fatalf("SetReadDeadline error %v", err)
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

		form.Username = c.username
		var buf bytes.Buffer

		err = tmpl.ExecuteTemplate(&buf, "message", form)
		if err != nil {
			fmt.Printf("failed to render room.tmpl: %v\n", err)
			continue
		}

		time := time.Now()

		c.Hub.broadcast <- buf.Bytes()

		log.Printf("ReadPump sending to channel: %p", msg)

		msg <- Message{
			RoomCode:   c.roomcode,
			Username:   c.username,
			Message:    form.Message,
			Time_Stamp: time,
		}
		log.Println("ReadPump successfully queued message")
	}
}

func (c *Client) WritePump() {
	Tick := time.NewTicker(PingPeriod)
	defer func() {
		Tick.Stop()
		err := c.conn.Close()
		if err != nil {
			log.Fatalf("write close %v", err)
		}
	}()

	for {
		select {
		case Message, ok := <-c.send:
			err := c.conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err != nil {
				log.Fatalf("writepump setwritedeadline %v", err)
			}
			if !ok {
				err := c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					log.Fatalf("writemessage %v", err)
				}
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				fmt.Printf("%v", err)
				return
			}
			_, err = w.Write(Message)
			if err != nil {
				log.Fatalf("failed to write message %v", err)
			}
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, err := w.Write(newline)
				if err != nil {
					log.Fatalf("writeLine %v", err)
				}
				_, err1 := w.Write(<-c.send)
				if err1 != nil {
					log.Fatalf("writeLine1 %v", err)
				}
			}

			if err := w.Close(); err != nil {
				fmt.Printf("%v", err)
				return
			}
		case <-Tick.C:
			err := c.conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err != nil {
				log.Fatalf("idk man the last one or smtg %v", err)
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				fmt.Printf("%v", err)
				return
			}
		}
	}
}
