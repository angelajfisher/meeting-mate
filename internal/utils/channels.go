package utils

var DataListeners = make(map[string]map[string]chan EventData) // map[meetingID]map[guildID]

func ReceiveZoomData(meetingID string, guildID string) <-chan EventData {
	c := make(chan EventData)

	if _, exists := DataListeners[meetingID]; !exists {
		DataListeners[meetingID] = make(map[string]chan EventData)
	}
	DataListeners[meetingID][guildID] = c

	return c
}
