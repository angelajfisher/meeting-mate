package utils

import "sync"

type dataListeners struct {
	listeners map[string]map[string]chan EventData // map[meetingID]map[guildID]
	mu        sync.RWMutex
}

func NewDataListeners() *dataListeners {
	return &dataListeners{
		listeners: make(map[string]map[string]chan EventData),
	}
}

// TODO: sanity checks so that existing channels aren't replaced
func (dl *dataListeners) AddListener(guildID string, meetingID string) chan EventData {
	c := make(chan EventData)

	dl.mu.Lock()
	defer dl.mu.Unlock()

	if _, exists := DataListeners.listeners[meetingID]; !exists {
		dl.listeners[meetingID] = make(map[string]chan EventData)
	}

	dl.listeners[meetingID][guildID] = c

	return c
}

func (dl *dataListeners) RemoveListener(guildID string, meetingID string, reason EventData) {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	dl.listeners[meetingID][guildID] <- reason
	close(dl.listeners[meetingID][guildID])
	delete(dl.listeners[meetingID], guildID)
}

func (dl *dataListeners) GetMeetingListeners(meetingID string) map[string]chan EventData {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	return dl.listeners[meetingID]
}
