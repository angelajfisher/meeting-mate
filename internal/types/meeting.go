package types

import (
	"sync"
)

type Meeting struct {
	Participants *ParticipantList
	name         string
	id           string
	startTime    string // unused
	endTime      string // unused
	mu           sync.RWMutex
}

type MeetingStore struct {
	Meetings map[string]*Meeting // map[meetingID]Meeting
	mu       sync.RWMutex
}

func newMeetingStore() *MeetingStore {
	return &MeetingStore{
		Meetings: make(map[string]*Meeting),
	}
}

func (ms *MeetingStore) NewMeeting(id string) *Meeting {
	if ms.exists(id) {
		return ms.Meetings[id]
	}

	newMeeting := &Meeting{
		id:           id,
		name:         "",
		startTime:    "",
		endTime:      "",
		Participants: newParticipantList(),
	}
	ms.add(newMeeting)
	return newMeeting
}

func (ms *MeetingStore) exists(meetingID string) bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if _, exists := ms.Meetings[meetingID]; exists {
		return true
	}
	return false
}

func (ms *MeetingStore) add(meeting *Meeting) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.Meetings[meeting.id] = meeting
}

func (m *Meeting) GetID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.id
}

// this is gross -- fix later
// names shouldn't/can't change for an ongoing meeting anyway, so this method shouldn't exist...
// just take the name on meeting creation
func (m *Meeting) Name(name string) string {
	m.mu.RLock()
	if m.name == name {
		m.mu.RUnlock()
		return name
	}

	m.mu.RUnlock()
	m.mu.Lock()
	defer m.mu.Unlock()

	m.name = name
	return name
}
