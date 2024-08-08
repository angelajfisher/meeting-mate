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

type Payload struct {
	PlainToken string `json:"plainToken"`
}

type Verification struct {
	PlainToken     string `json:"plainToken"`
	EncryptedToken string `json:"encryptedToken"`
}
type ZoomData struct {
	Payload Payload `json:"payload,omitempty"`
	EventTS int64   `json:"event_ts,omitempty"`
	Event   string  `json:"event,omitempty"`
}

func handleWebhooks(w http.ResponseWriter, r *http.Request) {
	// TODO: Check incoming event type and respond accordingly

	log.Println("Webhook received!")

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(string(reqBody))

	var data ZoomData
	err = json.Unmarshal(reqBody, &data)
	if err != nil {
		log.Println(err)
		return
	}

	hasher := hmac.New(sha256.New, []byte(Secret))
	hasher.Write([]byte(data.Payload.PlainToken))

	verification := Verification{PlainToken: data.Payload.PlainToken, EncryptedToken: hex.EncodeToString(hasher.Sum(nil))}

	retBody, err := json.Marshal(verification)
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
