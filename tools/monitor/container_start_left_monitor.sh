#!/bin/bash

if  [ ! "$1" ] ;then
    gap=5
else
    gap=${1}
fi


echo "Start Container Start Left Monitor,Check Gap ${1} second"

while :
do
    echo  "time:"`date`";left:"`sudo docker ps -a | grep Created | wc | awk '{print $1}'`
    sleep(${gap})
done