package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

var Secret string

const (
	EndpointValidation = "endpoint.url_validation"
)

type ZoomData struct {
	Payload interface{} `json:"payload"`
	EventTS int64       `json:"event_ts"`
	Event   string      `json:"event"`
}

type URLValidation struct {
	PlainToken     string `json:"plainToken"`
	EncryptedToken string `json:"encryptedToken,omitempty"`
}

func handleWebhooks(w http.ResponseWriter, r *http.Request) {

	log.Println("Webhook received!")

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(string(reqBody))

	var payload json.RawMessage
	data := ZoomData{Payload: &payload}
	err = json.Unmarshal(reqBody, &data)
	if err != nil {
		log.Println(err)
		return
	}

	switch data.Event {
	case EndpointValidation:
		var eventData URLValidation
		err = json.Unmarshal([]byte(payload), &eventData)
		if err != nil {
			log.Println(err)
		}

		hasher := hmac.New(sha256.New, []byte(Secret))
		hasher.Write([]byte(eventData.PlainToken))

		eventData.EncryptedToken = hex.EncodeToString(hasher.Sum(nil))

		retBody, err := json.Marshal(eventData)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(string(retBody))

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(retBody)
		if err != nil {
			log.Println(err)
			return
		}
	}

}
