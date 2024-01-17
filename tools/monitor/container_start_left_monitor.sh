#!/bin/bash
gap=1
echo "Start Container Start Left Monitor,Check Gap ${gap} second"
while :
do
    left=`sudo docker ps -a | grep Created | wc | awk '{print $1}'`

    if [ ${left} != "0" ]; then
        break
    fi
done
while :
do
    left=`sudo docker ps -a | grep Created | wc | awk '{print $1}'`
    echo  "time:"`date`";left:${left}"

    if [ ${left} == "0" ]; then
        break
    fi

    sleep ${gap}

done
