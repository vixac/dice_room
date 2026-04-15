package main

import (
	"dice_room/store"
	"dice_room/store/bullet_store"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/vixac/bullet/store/store_interface"

	"github.com/vixac/firbolg_clients/bullet/local_bullet"
	"github.com/vixac/firbolg_clients/bullet/rest_bullet"
)

func buildMemoryStore() *store.MemoryStore {
	fmt.Printf("Building memory store")
	return store.NewMemoryStore()
}

func buildBullet(bulletPort int) store.Store {
	fmt.Printf("Building bullet local")
	space := store_interface.TenancySpace{
		AppId:     5000,
		TenancyId: 0,
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	option := rest_bullet.WithLogger(logger)

	bulletStr := strconv.Itoa(bulletPort)

	restClient := rest_bullet.NewRestClient("http://localhost:"+bulletStr, space, option)

	fmt.Printf("VX: rest client %s\n", restClient.AppId)

	store := bullet_store.NewBulletStore(restClient)
	return store
}

func main() {

	space := store_interface.TenancySpace{
		AppId:     1234,
		TenancyId: 0,
	}
	fmt.Printf("Dice Room begins...\n")
	args, err := ReadArgs()
	if err != nil {
		log.Fatal("Error parsing args: ", err)
	}

	localBullet := local_bullet.LocalBullet{
		Space: space,
	}
	fmt.Printf("VX:local bullet %s\n", localBullet.Space)

	//ene, err := grove_engine.NewGroveEngine(&localBullet)
	//ene, err := grove_engine.NewGroveEngine(restClient)

	//
	broadcaster := NewBroadcaster()

	store := buildBullet(args.BulletPort)
	srv := NewServer(store, broadcaster, args.HostPrefix)

	addr := ":" + strconv.Itoa(args.Port)
	log.Println("Listening on " + addr)
	fmt.Println("Dice room is ready.")
	log.Fatal(http.ListenAndServe(addr, srv.routes()))
}
