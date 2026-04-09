package store

import (
	"dice_room/model"
	"strconv"
	"sync"
	"time"
)

// MemoryStore is the in-memory implementation of Store.
type MemoryStore struct {
	mu    sync.Mutex
	rooms map[string]*model.Room
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{rooms: make(map[string]*model.Room)}
}

func (s *MemoryStore) CreateRoom(name string) (*model.Room, error) {
	id := strconv.FormatInt(time.Now().UnixNano(), 36)
	if name == "" {
		name = id
	}
	room := &model.Room{ID: id, RoomName: name}
	s.mu.Lock()
	s.rooms[id] = room
	s.mu.Unlock()
	return room, nil
}

func (s *MemoryStore) GetRoom(id string) (*model.Room, error) {
	s.mu.Lock()
	room, ok := s.rooms[id]
	s.mu.Unlock()
	if !ok {
		return nil, ErrRoomNotFound
	}
	return room, nil
}

func (s *MemoryStore) AddEntry(roomID string, entry model.LogEntry) error {
	room, err := s.GetRoom(roomID)
	if err != nil {
		return err
	}
	room.Lock.Lock()
	room.Log = append(room.Log, entry)
	room.Lock.Unlock()
	return nil
}
