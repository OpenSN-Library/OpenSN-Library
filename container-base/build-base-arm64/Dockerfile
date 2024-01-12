FROM ubuntu:22.04

ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 以root用户权限进行登陆
USER root 
COPY materials/sources.list /etc/apt/sources.list 
RUN echo "net.ipv4.ip_forward = 1" >> /etc/sysctl.conf && sysctl -p
RUN apt update && apt install -y iproute2 \
    frr \
    traceroute \
    tcpdump \
    vim \
    iputils-ping \
    traceroute \
    iperf3
COPY scripts/entrypoint /Workspace/entrypoint
RUN chmod +x /Workspace/entrypoint
CMD ["bash", "/Workspace/entrypoint"]
