package main

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
)

type Args struct {
	Port       int
	BulletPort int
	HostPrefix string
}

func ReadArgs() (*Args, error) {
	var args Args
	internalBulletPort := flag.String("internalBulletPort", "", "port number to reach internal bullet")
	port := flag.String("port", "", "port number to run on")
	hostPrefix := flag.String("hostPrefix", "", "the /tbc/dice_room component of the url which is needed because firbolg_gateway trims it down.")

	flag.Parse()
	if *internalBulletPort == "" {
		return nil, errors.New("missing internal bullet port")
	}

	fmt.Println("hostPrefix (used for redirects to /room) for this session is " + *hostPrefix)

	if *port == "" {
		return nil, errors.New("missing port")
	}
	internalBulletPortInt, err := strconv.Atoi(*internalBulletPort)

	if err != nil {
		fmt.Println("Invalid internalBulletPort port :", internalBulletPort)
		return nil, err
	}
	fmt.Println("Bullet port (unused is " + strconv.Itoa(internalBulletPortInt))

	portInt, err := strconv.Atoi(*port)
	if err != nil {
		fmt.Println("Invalid port :", port)
		return nil, err
	}
	args.Port = portInt
	args.BulletPort = internalBulletPortInt
	args.HostPrefix = *hostPrefix
	return &args, nil
}
