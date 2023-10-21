FROM ubuntu_python:latest

USER root

# 添加run.bash
ADD run.bash /etc

# 拷贝satellite-node
COPY ./satellite_node /satellite_node

# 将netlink_test_userspace 进行拷贝
COPY ./netlink_test_userspace /netlink_test_userspace

RUN cd /netlink_test_userspace && cmake -S . -B build

RUN cd /netlink_test_userspace/build && make

# zhf add code
COPY ./udp_client_and_server_python /udp_client_and_server_python

# zhf add code
COPY ./tcp_client_and_server /tcp_client_and_server

# zhf add code
COPY ./udp_live_stream /udp_live_stream

COPY video_trans /video_trans

# 进入satellite-node目录
RUN cd /satellite_node

# 给bootstrap.sh加上可执行的权限
RUN chmod +x satellite_node/bootstrap.sh

# 运行satellite-node
CMD /bin/bash /etc/run.bash