FROM quay.io/coreos/etcd:v3.5.11-arm64
ENV ETCD_ADVERTISE_CLIENT_URLS=http://0.0.0.0:2379
ENV ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379
ENV ETCD_LISTEN_PEER_URLS=http://0.0.0.0:2380 
ENV ETCD_INITIAL_ADVERTISE_PEER_URLS=http://0.0.0.0:2380 
ENV ALLOW_NONE_AUTHENTICATION=yes 
ENV ETCD_INITIAL_CLUSTER=emulator_etcd=http://0.0.0.0:2380 
ENV ETCD_NAME=emulator_etcd
ENV ETCD_DATA_DIR=/opt/bitnami/etcd/data 
EXPOSE 2379
