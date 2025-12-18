package main

import (
	"log"
	"time"
)

func monitorStats() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		totalMsg, activeClients, totalClients := chatRoom.GetStats()

		log.Printf("Stats: %d active clients | %d total clients | %d messages", activeClients, totalClients, totalMsg)
	}
}
