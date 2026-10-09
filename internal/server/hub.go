package server

func NewHub() *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		broadcast:     make(chan Event),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		messageSender: make(map[string]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)

				for id, sender := range h.messageSender {
					if sender == client {
						delete(h.messageSender, id)
					}
				}
			}

		case event := <-h.broadcast:
			if event.Type == EventDelivered || event.Type == EventRead {
				sender, ok := h.messageSender[event.MessageID]
				if !ok || sender.roomcode != event.RoomCode {
					continue
				}

				select {
				case sender.send <- event:
				default:
				}

				continue
			}

			if event.Type == "text" && event.MessageID != "" && event.Sender != nil {
				h.messageSender[event.MessageID] = event.Sender
			}

			for client := range h.clients {
				if client.roomcode != event.RoomCode {
					continue
				}

				select {
				case client.send <- event:
				default:
					delete(h.clients, client)
					close(client.send)
				}
			}

			if event.Type == "text" && event.Sender != nil {
				select {
				case event.Sender.send <- Event{
					Type:      EventSent,
					MessageID: event.MessageID,
				}:
				default:
				}
			}
		}
	}
}
