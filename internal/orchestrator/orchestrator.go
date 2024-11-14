package orchestrator

import (
	"log"
	"time"

	"github.com/angelajfisher/meeting-mate/internal/types"
)

type Orchestrator struct {
	meetingWatches *types.Bimap // Bidirectional map tracking ongoing watches categorized by meetingID and by guildID
	dataListeners  *types.DataListeners
	allMeetings    *types.MeetingStore
}

func NewOrchestrator() Orchestrator {
	return Orchestrator{
		meetingWatches: types.NewBimap(),
		dataListeners:  types.NewDataListeners(),
		allMeetings:    types.NewMeetingStore(),
	}
}

// Whether the given meeting is being monitored by the system
func (o Orchestrator) IsWatchedMeeting(meetingID string) bool {
	return o.meetingWatches.ActiveMeeting(meetingID)
}

// Whether the given meeting has an ongoing watch in the given guild
func (o Orchestrator) IsOngoingWatch(guildID string, meetingID string) bool {
	return o.meetingWatches.Exists(guildID, meetingID)
}

// Lists all meetings being watched by a given guild
func (o Orchestrator) GetGuildMeetings(guildID string) []string {
	return o.meetingWatches.GetMeetings(guildID)
}

func (o Orchestrator) StartWatch(guildID string, meetingID string) <-chan types.UpdateData {
	o.allMeetings.NewMeeting(meetingID)
	o.meetingWatches.Add(guildID, meetingID)
	return o.dataListeners.Listen(guildID, meetingID)
}

func (o Orchestrator) UpdateMeeting(meetingID string, data types.MeetingData) {
	update := types.UpdateData{
		EventType:   data.EventType,
		MeetingName: data.MeetingName,
	}

	switch data.EventType {
	case types.ZoomParticipantJoin:
		update.Participants = o.allMeetings.AddParticipant(meetingID, data.ParticipantID, data.ParticipantName)
	case types.ZoomParticipantLeave:
		update.Participants = o.allMeetings.RemoveParticipant(meetingID, data.ParticipantID, data.ParticipantName)
	case types.ZoomMeetingEnd:
		update.MeetingDuration = calcMeetingDuration(data.StartTime, data.EndTime)
		update.TotalParticipants = o.allMeetings.EndMeeting(meetingID)
	}

	for _, dataChannel := range o.dataListeners.GetMeetingListeners(meetingID) {
		dataChannel <- update
	}
}

// Informs a watch process of a cancellation request so it can gracefully stop
func (o Orchestrator) CancelWatch(guildID string, meetingID string) {
	o.dataListeners.Remove(guildID, meetingID, types.UpdateData{EventType: types.WatchCanceled})
	o.meetingWatches.Remove(guildID, meetingID)
}

// Informs all watch processes of impeding shutdown so they can act accordingly
func (o Orchestrator) Shutdown() {
	for _, watch := range o.dataListeners.GetAllListeners() {
		watch <- types.UpdateData{EventType: types.BotShutdown}
	}
}

func calcMeetingDuration(start string, end string) string {
	calcDuration := true // whether to return actual calculation; changes to false upon error

	startTime, err := time.Parse(types.ZoomTimeFormat, start)
	if err != nil {
		log.Printf("could not parse meeting start time: %s", err)
		calcDuration = false
	}
	endTime, err := time.Parse(types.ZoomTimeFormat, end)
	if err != nil {
		log.Printf("could not parse meeting end time: %s", err)
		calcDuration = false
	}

	if calcDuration {
		return endTime.Sub(startTime).String()
	}
	return "Unknown"
}
