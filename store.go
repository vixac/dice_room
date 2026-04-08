package main

import (
	"errors"
	"strconv"
	"sync"
	"time"
)

// ErrRoomNotFound is returned when a room ID does not exist.
var ErrRoomNotFound = errors.New("room not found")

// Store is the interface for all room persistence operations.
// Swap the in-memory implementation for a database one without touching handlers.
type Store interface {
	CreateRoom(name string) (*Room, error)
	GetRoom(id string) (*Room, error)
	AddEntry(roomID string, entry LogEntry) error
}

// MemoryStore is the in-memory implementation of Store.
type MemoryStore struct {
	mu    sync.Mutex
	rooms map[string]*Room
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{rooms: make(map[string]*Room)}
}

func (s *MemoryStore) CreateRoom(name string) (*Room, error) {
	id := strconv.FormatInt(time.Now().UnixNano(), 36)
	if name == "" {
		name = id
	}
	room := &Room{ID: id, RoomName: name}
	s.mu.Lock()
	s.rooms[id] = room
	s.mu.Unlock()
	return room, nil
}

func (s *MemoryStore) GetRoom(id string) (*Room, error) {
	s.mu.Lock()
	room, ok := s.rooms[id]
	s.mu.Unlock()
	if !ok {
		return nil, ErrRoomNotFound
	}
	return room, nil
}

func (s *MemoryStore) AddEntry(roomID string, entry LogEntry) error {
	room, err := s.GetRoom(roomID)
	if err != nil {
		return err
	}
	room.Lock.Lock()
	room.Log = append(room.Log, entry)
	room.Lock.Unlock()
	return nil
}
