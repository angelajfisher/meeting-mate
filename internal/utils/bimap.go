package utils

import "log"

type bimap struct {
	guildMeetings map[string]map[string]struct{} // map[guildID]map[meetingID] - the meetings being watched by a guild
	meetingGuilds map[string]map[string]struct{} // map[meetingID]map[guildID] - the guilds watching a meeting
}

func newBimap() *bimap {
	return &bimap{
		guildMeetings: make(map[string]map[string]struct{}),
		meetingGuilds: make(map[string]map[string]struct{}),
	}
}

func (b *bimap) Add(guildID string, meetingID string) {
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

func (b *bimap) Remove(guildID string, meetingID string) {
	// Remove from guildMeetings
	if meetingList, exists := b.guildMeetings[guildID]; exists {
		if _, present := meetingList[meetingID]; present {
			log.Println("removing from meeting list")
			delete(b.guildMeetings[guildID], meetingID)
		}
	}

	// Remove from meetingGuilds
	if guildList, exists := b.meetingGuilds[guildID]; exists {
		if _, present := guildList[guildID]; present {
			log.Println("removing from guild list")
			delete(b.meetingGuilds[meetingID], guildID)
		}
	}
}

func (b *bimap) GetGuilds(meetingID string) map[string]struct{} {
	if guildList, exists := b.meetingGuilds[meetingID]; exists {
		return guildList
	}
	return make(map[string]struct{})
}

func (b *bimap) GetMeetings(guildID string) map[string]struct{} {
	if meetingList, exists := b.guildMeetings[guildID]; exists {
		return meetingList
	}
	return make(map[string]struct{})
}

func (b *bimap) Exists(guildID string, meetingID string) bool {
	if guildList, exists := b.guildMeetings[guildID]; exists {
		_, present := guildList[meetingID]
		return present
	}
	return false
}

func (b *bimap) ActiveMeeting(meetingID string) bool {
	log.Println(b.meetingGuilds[meetingID])
	if _, exists := b.meetingGuilds[meetingID]; exists && len(b.meetingGuilds[meetingID]) != 0 {
		return true
	}
	return false
}
