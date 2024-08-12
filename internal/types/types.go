package types

import "time"

type EventData struct {
	EventType       string
	MeetingName     string
	ParticipantName string
	ParticipantID   string
}

const (
	EndpointValidation = "endpoint.url_validation"
	MeetingEnd         = "meeting.ended"
	ParticipantJoin    = "meeting.participant_joined"
	ParticipantLeave   = "meeting.participant_left"
	Canceled           = "canceled"
	Shutdown           = "shutdown"
)

var (
	MeetingData    = make(chan EventData, 5)
	WatchMeetingID = make(chan string)
)

func CurrentTime() string {

	format := "2006-01-02 15:04:05"
	return time.Now().Format(format)

}
