package bullet_store

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/vixac/firbolg_clients/bullet/bullet_interface"
	bullet_stl "github.com/vixac/firbolg_clients/bullet/bullet_stl/containers"
)

type RoomCollection struct {
	Codec      Codec[RoomInfo]
	Collection bullet_stl.Collection
}

type RoomInfo struct {
	Id   string
	Name string
}
type RoomId struct {
	Id string
}

//VX:TODO room probably wants the same incrementing value. Same as got.

// for storing room info which is basically nothing for now.
func NewRoomCollection(bucketId int32, client bullet_interface.BulletClientInterface, codec Codec[RoomInfo]) RoomCollection {
	coll := bullet_stl.NewBulletCollection(bucketId, client, client)
	return RoomCollection{
		Collection: coll,
		Codec:      codec,
	}
}

// VX:TODO with this implementation room ids need to be unique.
func (r *RoomCollection) CreateRoom(name string) (*RoomId, error) {
	id := strconv.FormatInt(time.Now().UnixNano(), 36)
	if name == "" {
		name = id
	}
	now := time.Now()
	room := RoomInfo{
		Id:   id,
		Name: name,
	}
	encoded, err := r.Codec.Encode(room)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	fmt.Printf("VX: created room %s, %s\n", id, encoded)
	_, err = r.Collection.CreateItemUnder(id, encoded, &now)
	if err != nil {
		return nil, err
	}
	return &RoomId{Id: id}, nil
}

//VX:TODO split this out so that the roomifno is whats beign served up here.

// VX:TODO dont return room as rooms have logentires and this doesnt. Make a type
func (r *RoomCollection) GetRoom(id RoomId) (*RoomInfo, error) {

	items, err := r.Collection.ItemsForKeys([]string{id.Id})
	if err != nil {
		return nil, err
	}
	if len(items) != 1 {
		return nil, errors.New("wrong number of items for room")
	}

	for k, v := range items {
		fmt.Printf("VX: KEY IS %s, payload %s\n", k.Key, v.Payload)
		if id.Id != k.Key {
			return nil, errors.New("wrong room fetched")
		}
		var info RoomInfo
		err := r.Codec.Decode(v.Payload, &info)
		if err != nil {
			return nil, err
		}
		fmt.Printf("VX: info is %+v", info)
		return &info, err
	}
	return nil, errors.New("Room not found")
}
