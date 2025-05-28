package server

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/angelajfisher/meeting-mate/internal/types"
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

func (s Config) handleWebhooks(w http.ResponseWriter, r *http.Request) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}

	var payload json.RawMessage
	zoomData := ZoomData{Payload: &payload}
	err = json.Unmarshal(reqBody, &zoomData)
	if err != nil {
		log.Println(err)
		return
	}

	if zoomData.Event == types.ZOOM_ENDPOINT_VALIDATION {
		log.Println("Webhook received: URL validation request")

		var response []byte
		response, err = validateEndpoint(payload, s.Secret)
		if err != nil {
			log.Println(err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(response)
		if err != nil {
			log.Println(err)
			return
		}

		return
	}

	log.Println("Received webhook from " + r.Host + ": updating applicable watched meetings")

	var payloadData Meeting
	err = json.Unmarshal(payload, &ObjectWrapper{&payloadData})
	if err != nil {
		log.Println(err)
	}

	if !s.Orchestrator.IsWatchedMeeting(payloadData.ID) {
		return
	}

	// Will determine how we handle this data: do we forward it to the sister server, or were we sent this to sync up?
	synchronizing := r.Host == s.SisterAddress

	updatedMeetingData := types.MeetingData{
		EventType:   zoomData.Event,
		MeetingName: payloadData.Topic,
		Silent:      synchronizing,
	}

	if zoomData.Event == types.ZOOM_PARTICIPANT_JOIN || zoomData.Event == types.ZOOM_PARTICIPANT_LEAVE {
		updatedMeetingData.ParticipantName = payloadData.Participant.UserName
		updatedMeetingData.ParticipantID = payloadData.Participant.UserID
	} else if zoomData.Event == types.ZOOM_MEETING_END {
		updatedMeetingData.StartTime = payloadData.StartTime
		updatedMeetingData.EndTime = payloadData.EndTime
	}

	s.Orchestrator.UpdateMeeting(payloadData.ID, updatedMeetingData)

	// Quietly forward this data to our sister server if applicable
	if !synchronizing && s.SisterAddress != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		var req *http.Request
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, s.SisterAddress+WEBHOOK_SLUG, r.Body)
		if err != nil {
			fmt.Println("could not forward update:", err)
			return
		}

		req.Header.Add("content-type", "application/json")
		resp, reqErr := http.DefaultClient.Do(req)
		resp.Body.Close()
		if reqErr != nil {
			fmt.Println("Could not forward update data to sister server:", reqErr)
		} else {
			fmt.Println("Successfully forwarded update data to sister server")
		}
	}
}

func validateEndpoint(payload json.RawMessage, secret string) ([]byte, error) {
	var payloadData URLValidation
	err := json.Unmarshal(payload, &payloadData)
	if err != nil {
		return []byte{}, fmt.Errorf("could not validate Zoom endpoint: %w", err)
	}

	hasher := hmac.New(sha256.New, []byte(secret))
	hasher.Write([]byte(payloadData.PlainToken))

	payloadData.EncryptedToken = hex.EncodeToString(hasher.Sum(nil))

	var body []byte
	body, err = json.Marshal(payloadData)
	if err != nil {
		return []byte{}, fmt.Errorf("could not validate Zoom endpoint: %w", err)
	}

	return body, nil
}
