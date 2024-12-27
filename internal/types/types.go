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
	Flags             FeatureFlags
}

type FeatureFlags struct {
	Silent         bool   // Whether messages should be sent with the @silent flag
	JoinLink       string // User-supplied link for others to join the meeting
	Summaries      bool   // Whether meetings stats should be sent at the end of a meeting
	HistoryLevel   string // How many messages to send / delete as meetings start and end
	RestartCommand string // The command to restart this watch with the same flags
}

const (
	// Zoom event types
	ZOOM_ENDPOINT_VALIDATION = "endpoint.url_validation"
	ZOOM_MEETING_END         = "meeting.ended"
	ZOOM_PARTICIPANT_JOIN    = "meeting.participant_joined"
	ZOOM_PARTICIPANT_LEAVE   = "meeting.participant_left"

	// System notifications
	WATCH_CANCELED  = "canceled"
	SYSTEM_SHUTDOWN = "shutdown"
	UPDATE_FLAGS    = "update"

	// History level options -- MUST MATCH DATABASE SCHEMA
	FULL_HISTORY    = "Full"    // No old meeting messages are removed
	PARTIAL_HISTORY = "Partial" // Keep the old meeting message only if it's been buried by conversation
	MINIMAL_HISTORY = "Minimal" // Do not keep any old meeting messages

	ZOOM_TIME_FORMAT = "2006-01-02T15:04:05Z"
)
