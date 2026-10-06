package server

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

type Users struct {
	ID       int64  `bun:"id,pk,autoincrement"`
	Username string `bun:"username,unique,notnull" validate:"required"`
	Password string `bun:"password,notnull" validate:"required,cap,num,spec"`
}

type Rooms struct {
	RoomID   int64  `bun:"roomid,pk,autoincrement"`
	RoomName string `bun:"roomname" json:"roomname"`
	RoomCode string `bun:"roomcode,unique,notnull" json:"roomcode"`
}

type RoomMembers struct {
	ID       int64  `bun:"id,pk,autoincrement"`
	RoomName string `bun:"room_name"`
	RoomCode string `bun:"room_code"`
	Is_Admin bool   `bun:"role"`
	Username string `bun:"Username"`
	Room     *Rooms `bun:"rel:belongs-to,join:room_code=roomcode"`
	Users    *Users `bun:"rel:belongs-to,join:Username=username"`
}

type ToDisplay struct {
	RoomName string
	RoomCode string
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type PageData struct {
	Title    string
	RoomCode string
	RoomName string
	Roles    bool
}

type Client struct {
	Hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	username string
	roomcode string
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan Event
	register   chan *Client
	unregister chan *Client
}

type Message struct {
	RoomCode   string `json:"roomid"`
	Username   string `json:"username"`
	Message    string `json:"message"`
	Time_Stamp string `json:"time_stamp"`
}

type Event struct {
	Type      string `json:"type"`
	MessageID string `json:"message_id"`
	UserID    int64  `json:"user_id"`
	Content   string `json:"content"`
}
