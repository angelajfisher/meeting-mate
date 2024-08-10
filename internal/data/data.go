package data

type EventData struct {
	EventType       string
	MeetingName     string
	ParticipantName string
}

const (
	EndpointValidation = "endpoint.url_validation"
	MeetingStart       = "meeting.started"
	MeetingEnd         = "meeting.ended"
	ParticipantJoin    = "meeting.participant_joined"
	ParticipantLeave   = "meeting.participant_left"
)

var (
	MeetingData    = make(chan EventData)
	WatchMeetingID = make(chan string)
)
