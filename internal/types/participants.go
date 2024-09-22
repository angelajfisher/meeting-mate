package types

import (
	"strings"
	"sync"
)

type ParticipantList struct {
	participants map[string]string // map[participantID]participantName
	mu           sync.RWMutex
}

func newParticipantList() *ParticipantList {
	return &ParticipantList{
		participants: make(map[string]string),
	}
}

func (pl *ParticipantList) Add(participantID string, participantName string) {
	if name, _ := pl.exists(participantID); name == participantName {
		return
	}

	pl.mu.Lock()
	defer pl.mu.Unlock()

	pl.participants[participantID] = participantName
}

func (pl *ParticipantList) Remove(participantID string) {
	if _, exists := pl.exists(participantID); !exists {
		return
	}

	pl.mu.Lock()
	defer pl.mu.Unlock()

	delete(pl.participants, participantID)
}

func (pl *ParticipantList) Stringify() string {
	builder := new(strings.Builder)

	pl.mu.RLock()
	defer pl.mu.RUnlock()

	for _, name := range pl.participants {
		builder.WriteString(name + "\n")
	}
	if builder.String() == "" {
		builder.WriteString("Unknown")
	}

	return builder.String()
}

func (pl *ParticipantList) Empty() {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	clear(pl.participants)
}

func (pl *ParticipantList) exists(participantID string) (string, bool) {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	name, exists := pl.participants[participantID]
	return name, exists
}
