FROM ubuntu:20.04

LABEL mantainer="HeFan Zhang" description="Source routing"
# 设置时区
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
# 以root用户权限进行登陆
USER root
# 替换为国内源
# RUN sed -i 's/archive.ubuntu.com/mirrors.aliyun.com/g' /etc/apt/sources.list
COPY sources.list /etc/apt/sources.list
## Update the system and install required packages
RUN apt-get update 
RUN apt-get install -y \
    build-essential \
    zlib1g-dev \
    libncurses5-dev  \
    libgdbm-dev \
    libnss3-dev \
    libssl-dev \
    libreadline-dev \
    libffi-dev \
    wget \
    libnetfilter-queue-dev \
    net-tools \
    iptables \
    vim \
    iputils-ping \
    cmake \
    libsm6 \
    libxext6 \
    libxrender-dev \
    frr \
    xauth

RUN touch /root/.Xauthority
RUN xauth add unix:0.0 MIT-MAGIC-COOKIE-1  ed38891ce0eb40149fbdc76e11c94e76
# COPY the daemons to frr
COPY ./daemons /etc/frr/daemons

# 启用ipv4转发
RUN echo "net.ipv4.ip_forward = 1" >> /etc/sysctl.conf && sysctl -p