package types

import (
	"sync"
)

// Note: Potential refactor incoming to remove Meeting structs since they are currently unused

type Meeting struct {
	Participants *ParticipantList
	name         string
	id           string
	startTime    string // unused
	endTime      string // unused
	mu           sync.RWMutex
}

type MeetingStore struct {
	meetings map[string]*Meeting // map[meetingID]Meeting
	mu       sync.RWMutex
}

func NewMeetingStore() *MeetingStore {
	return &MeetingStore{
		meetings: make(map[string]*Meeting),
	}
}

func (ms *MeetingStore) NewMeeting(id string) *Meeting {
	if ms.exists(id) {
		return ms.meetings[id]
	}

	newMeeting := &Meeting{
		id:           id,
		name:         "",
		startTime:    "",
		endTime:      "",
		Participants: newParticipantList(),
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.meetings[id] = newMeeting
	return newMeeting
}

func (ms *MeetingStore) AddParticipant(meetingID string, participantID string, participantName string) string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	ms.meetings[meetingID].Participants.Add(participantID, participantName, true)
	return ms.meetings[meetingID].Participants.Stringify()
}

func (ms *MeetingStore) RemoveParticipant(meetingID string, participantID string, participantName string) string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	ms.meetings[meetingID].Participants.Remove(participantID, participantName)
	return ms.meetings[meetingID].Participants.Stringify()
}

func (ms *MeetingStore) EndMeeting(id string) int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return ms.meetings[id].Participants.Empty()
}

func (ms *MeetingStore) exists(meetingID string) bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if _, exists := ms.meetings[meetingID]; exists {
		return true
	}
	return false
}
