#!/bin/bash -er
set +e
# 删除以node开头的所有容器
# shellcheck disable=SC2046
# shellcheck disable=SC2006
# shellcheck disable=SC2062
# 暂停所有以node开头的容器
docker stop  `docker ps -a | grep Sat-* | awk '{print $1}'`
# 删除所有以node开头的容器
docker rm -f `docker ps -a | grep Sat-* | awk '{print $1}'`
# 停止satellite-monitor
docker stop  `docker ps -a | grep satellite | awk '{print $1}'`
# 删除所有satellite-monitor的容器
docker rm -f `docker ps -a | grep satellite | awk '{print $1}'`
docker stop  `docker ps -a | grep ground | awk '{print $1}'`
# 删除所有satellite-monitor的容器
docker rm -f `docker ps -a | grep ground | awk '{print $1}'`
docker network rm `docker network ls | grep Network | awk '{print $1}'`
# 删除所有以Network开头的网络
docker network rm `docker network ls | grep ground | awk '{print $1}'`
# 打印成功清除的信息
echo "All containers and networks have been removed."
