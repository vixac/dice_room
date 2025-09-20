#!/bin/bash
#
# go run . --internalBulletPort 10 --port 1234  --hostPrefix /tbc/dice_room
if [ -z "$1" ]
  then
        echo "You must provide a binary name"
        exit 1
fi
echo "Dice_room starting on $1 and we are in $(eval pwd)"
./$1 --internalBulletPort $BULLET_PORT --port $DICE_ROOM_PORT  --hostPrefix $DICE_ROOM_LOCATION
