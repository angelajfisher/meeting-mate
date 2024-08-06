package server

import (
	"log"
	"net/http"
)

var Secret string

func handleWebhooks(w http.ResponseWriter, r *http.Request) {

	log.Println("Webhook received!")

	// unmarshal json into zoom data struct
	// check event type
	// if event type is endpoint.url_validation, do the hashing stuff and send data back
	// see this for steps: https://developers.zoom.us/docs/api/rest/webhook-reference/#validate-your-webhook-endpoint
	// else, parse event data and forward to discord bot

}
