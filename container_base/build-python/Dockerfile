FROM ubuntu_modified:latest
## 将Python-3.6.9 拷贝解压缩
COPY Python-3.6.9.tgz /

# RUN tar -zxf Python-3.6.9.tgz
COPY Python-3.6.9.tgz /

RUN tar xvf Python-3.6.9.tgz
# 进入Python-3.6.9目录，安装Python-3.6.9
RUN cd Python-3.6.9 && ./configure && make && make install

RUN cd ..

## 将tgz包删除
RUN rm -rf Python-3.6.9.tgz

# 将python3 替换为 python
RUN ln -s /usr/local/bin/python3 /usr/bin/python

# 将pip3 替换为 pip
RUN ln -s /usr/local/bin/pip3 /usr/bin/pip

RUN pip config set global.index-url https://pypi.tuna.tsinghua.edu.cn/simple

# 拷贝配置文件
COPY requirements.txt /

RUN python -m pip install --upgrade pip

# 进行依赖的安装, 豆瓣源
RUN pip install -r requirements.txt --verbose