package bullet_store

import (
	"dice_room/model"
	"errors"
)

type RoomCollection struct {
	Client string
}

type RoomId struct {
	Id string
}

// for storing room info which is basically nothing for now.
func NewRoomCollection(client string) RoomCollection {
	return RoomCollection{}
}

func (r *RoomCollection) CreateRoom(id RoomId) error {
	return errors.New("not impl")
}

// VX:TODO dont return room as rooms have logentires and this doesnt. Make a type
func (r *RoomCollection) GetRoom(id RoomId) (*model.Room, error) {
	return nil, errors.New("not impl")
}
