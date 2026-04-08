package main

import "sync"

// Room holds the state for a single dice room.
type Room struct {
	ID       string
	RoomName string
	Log      []LogEntry
	Lock     sync.Mutex
}

// LogEntry is one roll event. JSON tags are used for SSE broadcasting.
type LogEntry struct {
	User   string `json:"user"`
	Dice   string `json:"dice"`
	Result int    `json:"result"`
	Desc   string `json:"desc,omitempty"`
	Time   string `json:"time"`
}

// RoomData is the view model passed to room.html.
type RoomData struct {
	ID           string
	RoomName     string
	Log          []LogEntry
	UserName     string
	SelectedDice string
	HostPrefix   string
}
