#!/bin/bash

pids=`ls /proc/ | grep -E "[0-9]+"`

for pid in ${pids}
do
    cat /proc/${pid}/status | grep -E "Name|switch"
done