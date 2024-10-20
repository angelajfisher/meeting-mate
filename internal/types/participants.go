package types

import (
	"strings"
	"sync"
)

type Participant struct {
	id      string
	name    string
	present bool
}

type ParticipantList struct {
	participants map[string]Participant // map[participantID]Participant
	mu           sync.RWMutex
}

func newParticipantList() *ParticipantList {
	return &ParticipantList{
		participants: make(map[string]Participant),
	}
}

func (pl *ParticipantList) Add(participantID string, participantName string, present bool) {
	if name, currentlyPresent := pl.present(participantID); name == participantName && currentlyPresent == present {
		return
	}

	pl.mu.Lock()
	defer pl.mu.Unlock()

	pl.participants[participantID] = Participant{id: participantID, name: participantName, present: present}
}

func (pl *ParticipantList) Remove(participantID string) {
	if _, exists := pl.present(participantID); !exists {
		pl.Add(participantID, "", false)
		return
	}

	pl.mu.Lock()
	defer pl.mu.Unlock()

	participant := pl.participants[participantID]
	participant.present = false
	pl.participants[participantID] = participant
}

func (pl *ParticipantList) Stringify() string {
	builder := new(strings.Builder)

	pl.mu.RLock()
	defer pl.mu.RUnlock()

	for _, participant := range pl.participants {
		if participant.present {
			builder.WriteString(participant.name + "\n")
		}
	}
	if builder.String() == "" {
		builder.WriteString("Unknown")
	}

	return builder.String()
}

func (pl *ParticipantList) Empty() int {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	numParticipants := len(pl.participants)
	clear(pl.participants)
	return numParticipants
}

func (pl *ParticipantList) present(participantID string) (string, bool) {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	participant, exists := pl.participants[participantID]
	if !exists {
		return "", false
	}
	return participant.name, participant.present
}
