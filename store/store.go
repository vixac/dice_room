package store

import (
	"dice_room/model"
	"errors"
)

// ErrRoomNotFound is returned when a room ID does not exist.
var ErrRoomNotFound = errors.New("room not found")

// Store is the interface for all room persistence operations.
// Swap the in-memory implementation for a database one without touching handlers.
type Store interface {
	CreateRoom(name string) (*model.Room, error)
	GetRoom(id string) (*model.Room, error)
	AddEntry(roomID string, entry model.LogEntry) error
}
