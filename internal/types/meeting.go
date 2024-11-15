package types

import (
	"log"
	"sync"
)

type Meeting struct {
	Participants *ParticipantList
	name         string
	id           string
	startTime    string // unused
	endTime      string // unused
}

type MeetingStore struct {
	meetings map[string]Meeting // map[meetingID]Meeting
	mu       sync.RWMutex
}

func NewMeetingStore() *MeetingStore {
	return &MeetingStore{
		meetings: make(map[string]Meeting),
	}
}

func (ms *MeetingStore) NewMeeting(id string) Meeting {
	if ms.exists(id) {
		return ms.meetings[id]
	}

	newMeeting := Meeting{
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

// Stores changes to meeting data. Currently only meeting "topics" (names) are tracked
func (ms *MeetingStore) UpdateMeeting(id string, updatedName string) {
	ms.mu.RLock()
	meeting, exists := ms.meetings[id]
	ms.mu.RUnlock()

	if exists {
		if meeting.name != updatedName {
			meeting.name = updatedName
			ms.mu.Lock()
			ms.meetings[id] = meeting
			ms.mu.Unlock()
		}
	} else {
		log.Println("could not update meeting: meeting id " + id + " doesn't exist")
	}
}

func (ms *MeetingStore) GetName(id string) string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return ms.meetings[id].name
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
