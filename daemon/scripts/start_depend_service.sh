#!/bin/bash
set +e
sudo docker rm -f opensn_etcd
sudo docker rm -f opensn_influxdb
sudo docker rm -f opensn_codeserver
set -e
sudo docker run -d --rm --name=opensn_etcd -p 2379:2379 realssd/opensn_etcd
sudo docker run -d --rm --name=opensn_influxdb  -p 8086:8086 realssd/opensn_influxdb
sudo docker run -d --rm --name=opensn_codeserver -v `pwd`/../TopoConfigurators:/workspace --cpus=1 -m 1g -p 8079:8443 realssd/opensn_codeserver