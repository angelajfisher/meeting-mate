package types

import (
	"sync"
)

type DataListeners struct {
	listeners      map[string]map[string]chan UpdateData // map[meetingID]map[guildID]
	totalListeners int
	mu             sync.RWMutex
}

func NewDataListeners() *DataListeners {
	return &DataListeners{
		listeners: make(map[string]map[string]chan UpdateData),
	}
}

func (dl *DataListeners) Listen(guildID string, meetingID string) <-chan UpdateData {
	if dl.exists(guildID, meetingID) {
		return dl.listeners[meetingID][guildID]
	}

	c := make(chan UpdateData, 1)

	dl.mu.Lock()
	defer dl.mu.Unlock()

	if _, exists := dl.listeners[meetingID]; !exists {
		dl.listeners[meetingID] = make(map[string]chan UpdateData)
	}

	dl.listeners[meetingID][guildID] = c
	dl.totalListeners++

	return c
}

func (dl *DataListeners) Remove(guildID string, meetingID string, reason UpdateData) {
	if !dl.exists(guildID, meetingID) {
		return
	}

	dl.mu.Lock()
	defer dl.mu.Unlock()

	dl.listeners[meetingID][guildID] <- reason
	close(dl.listeners[meetingID][guildID])
	delete(dl.listeners[meetingID], guildID)
	dl.totalListeners--
}

func (dl *DataListeners) GetMeetingListeners(meetingID string) map[string]chan UpdateData {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	return dl.listeners[meetingID]
}

func (dl *DataListeners) GetAllListeners() []chan UpdateData {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	if dl.totalListeners == 0 {
		return []chan UpdateData{}
	}

	allListeners := make([]chan UpdateData, 0, dl.totalListeners)
	for _, listeners := range dl.listeners {
		for _, listener := range listeners {
			allListeners = append(allListeners, listener)
		}
	}

	return allListeners
}

func (dl *DataListeners) exists(guildID string, meetingID string) bool {
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
