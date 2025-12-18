package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

var chatRoom *ChatRoom

func serveHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Content  string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		req.Username = "Anonymous"
	}

	if req.Content == "" {
		http.Error(w, "Content cannot be empty", http.StatusBadRequest)
		return
	}

	chatRoom.AddMessage(req.Username, req.Content)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	clientID := uuid.New().String()
	client := &Client{
		ID:       clientID,
		Messages: make(chan Message, 10),
	}

	chatRoom.register <- client

	defer func() {
		chatRoom.unregister <- client
	}()

	history := chatRoom.GetHistory()
	for _, msg := range history {
		data, _ := json.Marshal(msg)

		fmt.Fprintf(w, "data: %s\n\n", data)

		w.(http.Flusher).Flush()
	}

	for {
		select {
		case msg, ok := <-client.Messages:
			if !ok {
				return
			}

			data, _ := json.Marshal(msg)
			fmt.Fprintf(w, "data: %s\n\n", data)
			w.(http.Flusher).Flush()

		case <-r.Context().Done():
			return
		}
	}
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	totalMsg, activeClients, totalClients := chatRoom.GetStats()

	stats := map[string]int{
		"total_messages": totalMsg,
		"active_clients": activeClients,
		"total_clients":  totalClients,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
