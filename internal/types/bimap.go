package types

import "sync"

type Bimap struct {
	guildMeetings map[string]map[string]struct{} // map[guildID]map[meetingID] - the meetings being watched by a guild
	meetingGuilds map[string]map[string]struct{} // map[meetingID]map[guildID] - the guilds watching a meeting
	mu            sync.RWMutex
}

func NewBimap() *Bimap {
	return &Bimap{
		guildMeetings: make(map[string]map[string]struct{}),
		meetingGuilds: make(map[string]map[string]struct{}),
	}
}

func (b *Bimap) Add(guildID string, meetingID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Add to guildMeetings
	if meetingList, exists := b.guildMeetings[guildID]; exists {
		if _, present := meetingList[meetingID]; !present {
			b.guildMeetings[guildID][meetingID] = struct{}{}
		}
	} else {
		b.guildMeetings[guildID] = make(map[string]struct{})
		b.guildMeetings[guildID][meetingID] = struct{}{}
	}

	// Add to meetingGuilds
	if guildList, exists := b.meetingGuilds[guildID]; exists {
		if _, present := guildList[guildID]; !present {
			b.meetingGuilds[meetingID][guildID] = struct{}{}
		}
	} else {
		b.meetingGuilds[meetingID] = make(map[string]struct{})
		b.meetingGuilds[meetingID][guildID] = struct{}{}
	}
}

func (b *Bimap) Remove(guildID string, meetingID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Remove from guildMeetings
	if meetingList, exists := b.guildMeetings[guildID]; exists {
		if _, present := meetingList[meetingID]; present {
			delete(b.guildMeetings[guildID], meetingID)
		}
	}

	// Remove from meetingGuilds
	if guildList, exists := b.meetingGuilds[meetingID]; exists {
		if _, present := guildList[guildID]; present {
			delete(b.meetingGuilds[meetingID], guildID)
		}
	}
}

func (b *Bimap) GetGuilds(meetingID string) []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if guildList, exists := b.meetingGuilds[meetingID]; exists {
		allGuilds := make([]string, 0, len(guildList))
		for guildID := range guildList {
			allGuilds = append(allGuilds, guildID)
		}
		return allGuilds
	}
	return []string{}
}

func (b *Bimap) GetMeetings(guildID string) []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if meetingList, exists := b.guildMeetings[guildID]; exists {
		allMeetings := make([]string, 0, len(meetingList))
		for meetingID := range meetingList {
			allMeetings = append(allMeetings, meetingID)
		}
		return allMeetings
	}
	return []string{}
}

func (b *Bimap) Exists(guildID string, meetingID string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if guildList, exists := b.guildMeetings[guildID]; exists {
		_, present := guildList[meetingID]
		return present
	}
	return false
}

func (b *Bimap) ActiveMeeting(meetingID string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if _, exists := b.meetingGuilds[meetingID]; exists && len(b.meetingGuilds[meetingID]) != 0 {
		return true
	}
	return false
}
