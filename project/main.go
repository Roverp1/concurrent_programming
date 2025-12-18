package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	chatRoom = NewChatRoom()

	go chatRoom.Run()

	go monitorStats()

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/post", handlePost)
	http.HandleFunc("/stream", handleStream)
	http.HandleFunc("/stats", handleStats)

	port := 6969
	fmt.Printf("Server starting on port: %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
