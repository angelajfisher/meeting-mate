package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/angelajfisher/zoom-bot/internal/types"
)

var Secret string

type ZoomData struct {
	Payload interface{} `json:"payload"`
	EventTS int64       `json:"event_ts"`
	Event   string      `json:"event"`
}

type URLValidation struct {
	PlainToken     string `json:"plainToken"`
	EncryptedToken string `json:"encryptedToken,omitempty"`
}

type Meeting struct {
	Duration    uint32      `json:"duration"`
	StartTime   string      `json:"start_time,omitempty"`
	EndTime     string      `json:"end_time,omitempty"`
	Timezone    string      `json:"timezone"`
	Topic       string      `json:"topic"`
	ID          string      `json:"id"`
	Type        uint8       `json:"type"`
	UUID        string      `json:"uuid"`
	HostID      string      `json:"host_id"`
	Participant Participant `json:"participant,omitempty"`
}

type Participant struct {
	UserID            string `json:"user_id"`
	UserName          string `json:"user_name"`
	ParticipantUserID string `json:"participant_user_id"`
	ID                string `json:"id"`
	JoinTime          string `json:"join_time,omitempty"`
	LeaveTime         string `json:"leave_time,omitempty"`
	Email             string `json:"email"`
	ParticipantUUID   string `json:"participant_uuid"`
	LeaveReason       string `json:"leave_reason,omitempty"`
}

// Wrapper to handle extracting nested data from Zoom webhook.
// Use with `json.Unmarshal(payload, &ObjectWrapper{&dataStruct})`
type ObjectWrapper struct {
	Object interface{} `json:"object"`
}

func handleWebhooks(w http.ResponseWriter, r *http.Request) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}

	var payload json.RawMessage
	eventData := ZoomData{Payload: &payload}
	err = json.Unmarshal(reqBody, &eventData)
	if err != nil {
		log.Println(err)
		return
	}

	var botData types.EventData

	switch {
	case eventData.Event == types.EndpointValidation:
		log.Println("Webhook received: URL validation request")

		var payloadData URLValidation
		err = json.Unmarshal(payload, &payloadData)
		if err != nil {
			log.Println(err)
		}

		hasher := hmac.New(sha256.New, []byte(Secret))
		hasher.Write([]byte(payloadData.PlainToken))

		payloadData.EncryptedToken = hex.EncodeToString(hasher.Sum(nil))

		var retBody []byte
		retBody, err = json.Marshal(payloadData)
		if err != nil {
			log.Println(err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(retBody)
		if err != nil {
			log.Println(err)
			return
		}
	case meetingID != "":
		log.Println("Webhook received: Updating watched meeting")

		var payloadData Meeting
		err = json.Unmarshal(payload, &ObjectWrapper{&payloadData})
		if err != nil {
			log.Println(err)
		}

		if payloadData.ID != meetingID {
			return
		}

		botData = types.EventData{EventType: eventData.Event, MeetingName: payloadData.Topic}

		if eventData.Event == types.ParticipantJoin || eventData.Event == types.ParticipantLeave {
			botData.ParticipantName = payloadData.Participant.UserName
			botData.ParticipantID = payloadData.Participant.UserID
		}

		types.MeetingData <- botData
	default:
		log.Println("Webhook received: No watched meeting to update")
		return
	}
}
