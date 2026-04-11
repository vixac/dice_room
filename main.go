package main

import (
	"dice_room/store/bullet_store"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	sqlite_store "github.com/vixac/bullet/store/sqlite"
	"github.com/vixac/bullet/store/store_interface"
	"github.com/vixac/firbolg_clients/bullet/local_bullet"
)

func main() {

	space := store_interface.TenancySpace{
		AppId:     123,
		TenancyId: 0,
	}
	path := os.Getenv("GOT_BOLT")
	if path == "" {
		log.Fatal("missing env GOT_BOLT, which should be the path to the got bolt file")
	}

	sqlitePath := os.Getenv("GOT_SQLITE")
	if sqlitePath == "" {
		log.Fatal("missing env GOT_SQLITE, which should be the path to the got sqlite file")
	}
	sqlite, err := sqlite_store.NewSQLiteStore(sqlitePath)
	if err != nil {
		log.Fatal(err)
	}

	localBullet := local_bullet.LocalBullet{
		Space: space,
		Store: sqlite,
	}

	fmt.Println("Dice Room begins... %+v\n", localBullet)
	args, err := ReadArgs()
	if err != nil {
		log.Fatal("Error parsing args: ", err)
	}

	//store := store.NewMemoryStore()
	store := bullet_store.NewBulletStore(&localBullet)
	broadcaster := NewBroadcaster()
	srv := NewServer(store, broadcaster, args.HostPrefix)

	addr := ":" + strconv.Itoa(args.Port)
	log.Println("Listening on " + addr)
	fmt.Println("Dice room is ready.")
	log.Fatal(http.ListenAndServe(addr, srv.routes()))
}
