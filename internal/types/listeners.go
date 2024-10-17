package types

import "sync"

type dataListeners struct {
	listeners map[string]map[string]chan EventData // map[meetingID]map[guildID]
	mu        sync.RWMutex
}

func newDataListeners() *dataListeners {
	return &dataListeners{
		listeners: make(map[string]map[string]chan EventData),
	}
}

func (dl *dataListeners) Listen(guildID string, meetingID string) chan EventData {
	if dl.exists(guildID, meetingID) {
		return dl.listeners[meetingID][guildID]
	}

	c := make(chan EventData, 1)

	dl.mu.Lock()
	defer dl.mu.Unlock()

	if _, exists := dl.listeners[meetingID]; !exists {
		dl.listeners[meetingID] = make(map[string]chan EventData)
	}

	dl.listeners[meetingID][guildID] = c

	return c
}

func (dl *dataListeners) Remove(guildID string, meetingID string, reason EventData) {
	if !dl.exists(guildID, meetingID) {
		return
	}

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

func (dl *dataListeners) exists(guildID string, meetingID string) bool {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	if _, exists := dl.listeners[meetingID]; !exists {
		return false
	}

	if _, exists := dl.listeners[meetingID][guildID]; !exists {
		return false
	}

	return true
}
