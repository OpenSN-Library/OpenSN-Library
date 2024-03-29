FROM lscr.io/linuxserver/code-server:latest

COPY materials/sources.list /etc/apt/sources.list
COPY materials/extensions /config/extensions

RUN /app/code-server/bin/code-server --install-extension /config/extensions/*


ENV SUDO_PASSWORD=opensn
ENV TZ=Asia/Shanghai
ENV PGID=1000
ENV PUID=1000
ENV DEFAULT_WORKSPACE=/workspace

RUN sudo apt update && sudo apt install -y \
    python3 \
    python3-pip \
    git \
    vim \
    tree 


RUN sudo pip3 install -i https://pypi.tuna.tsinghua.edu.cn/simple \
    ephem \
    etcd3 \
    loguru \
    requests 

RUN sudo pip3 install -i https://pypi.tuna.tsinghua.edu.cn/simple protobuf==3.20.0

VOLUME [ "/workspace" ]