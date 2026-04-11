package bullet_store

import (
	"dice_room/model"
	"dice_room/store"
	"fmt"

	"github.com/vixac/firbolg_clients/bullet/bullet_interface"
)

const (
	roomBuckId   int32 = 3000
	rollBucketId int32 = 3001
)

type BulletRoomStore struct {
	Client bullet_interface.BulletClientInterface //VX:TODO FirbolgClient
	Rooms  *RoomCollection
	Rolls  *RollCollection
}

func NewBulletStore(client bullet_interface.BulletClientInterface) store.Store {
	fmt.Printf("VX: Using bullet store\n.")

	roomCodec := JSONCodec[RoomInfo]{}
	rooms := NewRoomCollection(roomBuckId, client, &roomCodec)
	rollCodec := JSONCodec[model.LogEntry]{}
	rolls := NewRollCollection(rollBucketId, client, &rollCodec)
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

	id, err := b.Rooms.CreateRoom(name)

	if err != nil {
		return nil, err
	}
	return &model.Room{
		Id:       id.Id,
		RoomName: name,
	}, nil

}

func (b *BulletRoomStore) GetRoom(id string) (*model.Room, error) {
	roomId := roomIdFor(id)
	roomInfo, err := b.Rooms.GetRoom(roomId)
	if err != nil || roomInfo == nil {
		return nil, err
	}
	logs, err := b.Rolls.RollsForRoom(roomId) //VX:TODO paging one day.
	room := model.Room{
		Id:       roomInfo.Id,
		RoomName: roomInfo.Name,
		Log:      logs,
	}
	room.Log = logs
	return &room, err
}

func (b *BulletRoomStore) AddEntry(roomID string, entry model.LogEntry) error {
	return b.Rolls.AddRoll(roomIdFor(roomID), entry)
}
