package webhooks

import (
	"log"
	"net/http"
	"os"
)

var Port string
var Secret string

func Listen(devMode bool) {

	http.HandleFunc("/webhooks", handleWebhooks)
	log.Println("Zoom webhook listener starting on", Port)

	if devMode {
		err := http.ListenAndServe(Port, nil)
		if err != nil {
			log.Fatal("Could not start Zoom webhook listener:", err)
		}
	} else {
		err := http.ListenAndServeTLS(Port, os.Getenv("SSL_CERT"), os.Getenv("SSL_KEY"), nil)
		if err != nil {
			log.Fatal("Could not start Zoom webhook listener:", err)
		}
	}

}

func handleWebhooks(w http.ResponseWriter, r *http.Request) {

	log.Println("Webhook received!")

}
