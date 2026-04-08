package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func main() {
	fmt.Println("Dice Room begins...")
	args, err := ReadArgs()
	if err != nil {
		log.Fatal("Error parsing args: ", err)
	}

	store := NewMemoryStore()
	broadcaster := NewBroadcaster()
	srv := NewServer(store, broadcaster, args.HostPrefix)

	addr := ":" + strconv.Itoa(args.Port)
	log.Println("Listening on " + addr)
	fmt.Println("Dice room is ready.")
	log.Fatal(http.ListenAndServe(addr, srv.routes()))
}
