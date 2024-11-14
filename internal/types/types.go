package types

type MeetingData struct {
	EventType       string
	MeetingName     string
	ParticipantName string
	ParticipantID   string
	StartTime       string
	EndTime         string
}

type UpdateData struct {
	EventType         string
	MeetingName       string
	Participants      string
	TotalParticipants int
	MeetingDuration   string
}

const (
	ZoomEndpointValidation = "endpoint.url_validation"
	ZoomMeetingEnd         = "meeting.ended"
	ZoomParticipantJoin    = "meeting.participant_joined"
	ZoomParticipantLeave   = "meeting.participant_left"
	WatchCanceled          = "canceled"
	BotShutdown            = "shutdown"
	ZoomTimeFormat         = "2006-01-02T15:04:05Z"
)
