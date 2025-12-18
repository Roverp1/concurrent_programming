package main

import (
	"sync"
	"time"
)

type Message struct {
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type Client struct {
	ID       string
	Messages chan Message
}

type ChatRoom struct {
	clients    map[string]*Client
	mutex      sync.RWMutex
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	history    []Message
	historyMu  sync.Mutex
	stats      Stats
}

type Stats struct {
	TotalMessages int
	ActiveClients int
	TotalClients  int
	mu            sync.Mutex
}

func NewChatRoom() *ChatRoom {
	return &ChatRoom{
		clients:    make(map[string]*Client),
		broadcast:  make(chan Message, 100),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		history:    make([]Message, 0),
	}
}

func (cr *ChatRoom) Run() {
	for {
		select {
		case client := <-cr.register:
			cr.mutex.Lock()
			cr.clients[client.ID] = client
			cr.mutex.Unlock()

			cr.stats.mu.Lock()
			cr.stats.ActiveClients = len(cr.clients)
			cr.stats.TotalClients++
			cr.stats.mu.Unlock()

		case client := <-cr.unregister:
			cr.mutex.Lock()
			if _, ok := cr.clients[client.ID]; ok {
				delete(cr.clients, client.ID)
				close(client.Messages)
			}
			cr.mutex.Unlock()

			cr.stats.mu.Lock()
			cr.stats.ActiveClients = len(cr.clients)
			cr.stats.mu.Unlock()

		case message := <-cr.broadcast:
			cr.historyMu.Lock()
			cr.history = append(cr.history, message)
			if len(cr.history) > 100 {
				cr.history = cr.history[1:]
			}
			cr.historyMu.Unlock()

			cr.stats.mu.Lock()
			cr.stats.TotalMessages++
			cr.stats.mu.Unlock()

			cr.mutex.RLock()
			for _, client := range cr.clients {
				select {
				case client.Messages <- message:
				default:
				}
			}
			cr.mutex.RUnlock()
		}
	}
}

func (cr *ChatRoom) AddMessage(username, content string) {
	msg := Message{
		Username:  username,
		Content:   content,
		Timestamp: time.Now(),
	}
	cr.broadcast <- msg
}

func (cr *ChatRoom) GetHistory() []Message {
	cr.historyMu.Lock()
	defer cr.historyMu.Unlock()

	history := make([]Message, len(cr.history))
	copy(history, cr.history)
	return history
}

func (cr *ChatRoom) GetStats() (int, int, int) {
	cr.stats.mu.Lock()
	defer cr.stats.mu.Unlock()
	return cr.stats.TotalMessages, cr.stats.ActiveClients, cr.stats.TotalClients
}
