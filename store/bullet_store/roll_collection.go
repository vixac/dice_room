package bullet_store

import (
	"dice_room/model"
	"errors"
)

type RollCollection struct {
	Client string
}

// for storing room info which is basically nothing for now.
func NewRollCollection(client string) RollCollection {
	return RollCollection{}
}

func (r *RollCollection) AddRoll(room RoomId, roll model.LogEntry) error {
	return errors.New("not impl")
}

func (r *RollCollection) RollsForRoom(id RoomId) ([]model.LogEntry, error) {
	return nil, errors.New("not impl")
}
