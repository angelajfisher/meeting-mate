package webhooks

import (
	"fmt"
	"log"
	"net/http"
)

var Secret string

func Listen() {
	http.HandleFunc("/webhooks", handleWebhooks)
	fmt.Println("Zoom webhook listener starting on :12345")
	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		log.Fatal("Could not start Zoom webhook listener:", err)
	}
}

func handleWebhooks(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Webhook received!")
}
