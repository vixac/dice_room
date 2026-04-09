package bullet_store

import (
	"dice_room/model"
	"dice_room/store"
	"errors"
	"fmt"
)

type BulletRoomStore struct {
	Client string //VX:TODO FirbolgClient
	Rooms  *RoomCollection
	Rolls  *RollCollection
}

func NewBulletStore(client string) store.Store {
	fmt.Printf("VX: Using bullet store\n.")

	rooms := NewRoomCollection(client)
	rolls := NewRollCollection(client)
	return &BulletRoomStore{
		Client: client,
		Rooms:  &rooms,
		Rolls:  &rolls,
	}
}

func roomIdFor(name string) RoomId {
	return RoomId{
		Id: name,
	}
}
func (b *BulletRoomStore) CreateRoom(name string) (*model.Room, error) {
	//Room can use a collection for the payloads
	//and for the entries, we can just have another collection
	//that uses room id as a prefix. Easy.
	b.Rooms.CreateRoom(roomIdFor(name))
	return nil, errors.New("Not impl")
}

func (b *BulletRoomStore) GetRoom(id string) (*model.Room, error) {
	roomId := roomIdFor(id)
	room, err := b.Rooms.GetRoom(roomId)
	if err != nil || room == nil {
		return nil, err
	}
	logs, err := b.Rolls.RollsForRoom(roomId) //VX:TODO paging one day.
	room.Log = logs
	return room, err
}

func (b *BulletRoomStore) AddEntry(roomID string, entry model.LogEntry) error {
	return b.Rolls.AddRoll(roomIdFor(roomID), entry)
}
