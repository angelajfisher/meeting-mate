package utils

type EventData struct {
	EventType       string
	MeetingName     string
	ParticipantName string
	ParticipantID   string
}

const (
	ZoomEndpointValidation = "endpoint.url_validation"
	ZoomMeetingEnd         = "meeting.ended"
	ZoomParticipantJoin    = "meeting.participant_joined"
	ZoomParticipantLeave   = "meeting.participant_left"
	WatchCanceled          = "canceled"
	BotShutdown            = "shutdown"
)

var (
	// Bidirectional map tracking ongoing watches categorized by meetingID and by guildID
	MeetingWatches  = newBimap()
	UpdateMeetingID chan struct{ string bool }
)
