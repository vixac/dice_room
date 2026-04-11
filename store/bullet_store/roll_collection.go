package bullet_store

import (
	"dice_room/model"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/vixac/firbolg_clients/bullet/bullet_interface"
	bullet_stl "github.com/vixac/firbolg_clients/bullet/bullet_stl/containers"
	ids "github.com/vixac/firbolg_clients/bullet/bullet_stl/ids"
)

type LogId struct {
	EntryId     ids.BulletId
	RoomId      RoomId
	CreatedTime time.Time
}

func (k *LogId) Next(time time.Time) LogId {
	return LogId{
		RoomId:      k.RoomId,
		EntryId:     k.EntryId.Next(),
		CreatedTime: time,
	}
}

func FirstNoteId() ids.BulletId {
	id, _ := ids.NewBulletIdFromInt(828)
	return *id
}

func (k *LogId) ToString() string {
	createdStr := TimeToMillisString(k.CreatedTime)
	return k.RoomId.Id + ":" + k.EntryId.AasciValue + ":" + createdStr
}

type RollCollection struct {
	Codec      Codec[model.LogEntry]
	Collection bullet_stl.Collection
}

func FirstEntryId() ids.BulletId {
	id, _ := ids.NewBulletIdFromInt(828)
	return *id
}

// for storing room info which is basically nothing for now.
func NewRollCollection(bucketId int32, client bullet_interface.BulletClientInterface, codec Codec[model.LogEntry]) RollCollection {
	coll := bullet_stl.NewBulletCollection(bucketId, client, client)
	return RollCollection{
		Collection: coll,
		Codec:      codec,
	}
}

func NewLogIdFromString(input string) (*LogId, error) {
	split := strings.Split(input, ":")
	if len(split) != 3 {
		fmt.Printf("VX: this is not a valid logId: %s\n", input)
		return nil, errors.New("Invalid logId key")
	}
	roomInput := split[0]
	noteInput := split[1]
	createdMillisInput := split[2]
	roomId := RoomId{Id: roomInput}
	bulletId, err := ids.NewBulletIdFromString(noteInput)
	if err != nil {
		return nil, err
	}

	createdTime, err := EpochMillisStringToDate(createdMillisInput)
	if err != nil {
		return nil, err
	}
	longForm := LogId{
		RoomId:      roomId,
		EntryId:     *bulletId,
		CreatedTime: *createdTime,
	}
	return &longForm, nil
}

func highestIdInside(collection map[bullet_stl.CollectionId]bullet_stl.CollectionItem) (*ids.BulletId, error) {
	if len(collection) == 0 {
		return nil, nil
	}
	var highestIntValue int64 = 0
	for k, _ := range collection {

		logId, err := NewLogIdFromString(k.Key)
		if err != nil {
			return nil, err
		}
		if logId.EntryId.IntValue > highestIntValue {
			highestIntValue = logId.EntryId.IntValue
		}
	}
	return ids.NewBulletIdFromInt(highestIntValue)
}

func (r *RollCollection) NextIdForRoom(room RoomId, now time.Time) (*LogId, error) {
	existing, err := r.Collection.AllItemsUnderPrefix(room.Id)
	if err != nil {
		return nil, err
	}

	highestExistingId, err := highestIdInside(existing)
	if err != nil {
		return nil, err
	}
	if highestExistingId == nil { //this is the first note for this gotid
		first := FirstEntryId()
		highestExistingId = &first
	}

	newLogId := LogId{
		RoomId:      room,
		EntryId:     highestExistingId.Next(),
		CreatedTime: now,
	}
	return &newLogId, nil

}

func (r *RollCollection) AddRoll(room RoomId, roll model.LogEntry) error {
	now := time.Now()

	newId, err := r.NextIdForRoom(room, now)
	if err != nil {
		return err
	}
	newStringId := newId.ToString()

	encoded, err := r.Codec.Encode(roll)
	if err != nil {
		return err
	}
	_, err = r.Collection.CreateItemUnder(newStringId, encoded, &now)
	if err != nil {
		return err
	}
	return nil
}

// VX:TODO make this generic for the key and encode type, pass in the decoder or whatever.
func (r *RollCollection) collectionToLongFormMap(collection map[bullet_stl.CollectionId]bullet_stl.CollectionItem) (map[RoomId][]model.LogEntry, error) {
	idsToBlocks := make(map[RoomId][]model.LogEntry)
	for k, v := range collection {
		logId, err := NewLogIdFromString(k.Key)
		if err != nil || logId == nil {
			return nil, err
		}

		var entry model.LogEntry
		if err := r.Codec.Decode(v.Payload, &entry); err != nil {
			return nil, err
		}

		roomId := logId.RoomId
		existing, ok := idsToBlocks[roomId]
		if !ok {
			idsToBlocks[roomId] = []model.LogEntry{entry}
		} else {
			idsToBlocks[roomId] = append(existing, entry)
		}
	}
	return idsToBlocks, nil
}

func (r *RollCollection) RollsForRoom(id RoomId) ([]model.LogEntry, error) {

	res, err := r.Collection.AllItemsUnderPrefix(id.Id)
	if err != nil || len(res) == 0 {
		return nil, err
	}

	idMap, err := r.collectionToLongFormMap(res)
	if err != nil {
		return nil, err
	}
	//no entries for this id
	if len(idMap) == 0 {
		return nil, nil
	}
	if len(idMap) != 1 {
		return nil, errors.New("Too many rooms in response for logIds for id")
	}

	entries, ok := idMap[id]
	if !ok {
		return nil, errors.New("wrong id returned")
	}
	return entries, nil

}

func TimeToMillisString(time time.Time) string {
	millis := time.UnixMilli()
	return strconv.FormatInt(millis, 10)
}

func EpochMillisStringToDate(millisStr string) (*time.Time, error) {

	millis, err := strconv.ParseInt(millisStr, 10, 64)
	if err != nil {
		return nil, err
	}
	t := time.Unix(0, millis*int64(time.Millisecond))
	return &t, nil
}
