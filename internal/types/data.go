package types

type EventData struct {
	EventType       string
	MeetingName     string
	ParticipantName string
	ParticipantID   string
	StartTime       string
	EndTime         string
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

var (
	// Bidirectional map tracking ongoing watches categorized by meetingID and by guildID
	MeetingWatches  = newBimap()
	UpdateMeetingID chan struct{ string bool }
	DataListeners   = newDataListeners()
	AllMeetings     = newMeetingStore()
)
