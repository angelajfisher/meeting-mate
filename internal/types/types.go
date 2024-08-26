package types

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
