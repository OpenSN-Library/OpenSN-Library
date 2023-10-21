FROM ubuntu:20.04

# 该镜像仅为基础镜像，直接运行无任何功能，需要在其基础上进行扩展
# 镜像默认入口为/init.sh, 内置信号捕获, 可以加快容器停止与删除
# /init.sh 会调用/workspace/start.sh, 该脚本为容器启动脚本
# 请确保制作的镜像中包含/workspace/start.sh, 否则默认的脚本没有任何作用

LABEL mantainer="HeFan Zhang" description="Source routing"
# 设置时区
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 以root用户权限进行登陆
USER root 

# 替换为清华源
COPY materials/sources.list /etc/apt/sources.list 

# 安装依赖应用
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

# 配置x11转发
RUN touch /root/.Xauthority
RUN xauth add unix:0.0 MIT-MAGIC-COOKIE-1  ed38891ce0eb40149fbdc76e11c94e76

# 添加frr守护进程配置
COPY materials/daemons /etc/frr/daemons

# 启用ipv4转发
RUN echo "net.ipv4.ip_forward = 1" >> /etc/sysctl.conf && sysctl -p

# 从源代码安装Python3.6.9
COPY materials/Python-3.6.9.tgz /
RUN tar xvf Python-3.6.9.tgz
RUN cd Python-3.6.9 && ./configure && make && make install
RUN cd ..
RUN rm -rf Python-3.6.9.tgz
RUN ln -s /usr/local/bin/python3 /usr/bin/python
RUN ln -s /usr/local/bin/pip3 /usr/bin/pip
# 配置pypi 清华源
RUN pip config set global.index-url https://pypi.tuna.tsinghua.edu.cn/simple

# 安装pip依赖
COPY requirements.txt /
RUN python -m pip install --upgrade pip
RUN pip install -r requirements.txt --verbose

COPY container-scripts/init.sh /
COPY container-scripts/start.sh /workspace/start.sh
