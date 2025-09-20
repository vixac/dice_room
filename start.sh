#!/bin/bash
# 
# go run ./cmd/bullet -bolt boltdb -mongo $MONGO_PASS -db-type mongodb -port 10
if [ -z "$1" ]
  then
        echo "You must provide a binary name"
        exit 1
fi
echo "Dice_room starting on $1 and we are in $(eval pwd)"
./$1 --internalBulletPort $BULLET_PORT --port $DICE_ROOM_PORT
