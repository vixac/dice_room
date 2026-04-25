package model

import "sync"

// Room holds the state for a single dice room.
type Room struct {
	Id       string
	RoomName string
	Log      []LogEntry
	Lock     sync.Mutex
}

// LogEntry is one roll event. JSON tags are used for SSE broadcasting.
type LogEntry struct {
	User       string `json:"user"`
	Dice       string `json:"dice"`
	Result     int    `json:"result"`
	Desc       string `json:"desc,omitempty"`
	Time       string `json:"time"`
	UnixMillis int64  `json:"unixMillis"`
}

// PageData is the base view model passed to all templates.
type PageData struct {
	HostPrefix string
}

// RoomData is the view model passed to room.html.
type RoomData struct {
	PageData
	ID       string
	RoomName string
	Log      []LogEntry
	UserName string
}
