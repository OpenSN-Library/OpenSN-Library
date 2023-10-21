# 卫星网络模拟器-README

[TOC]

## 1. 使用

### 1.0 安装基础依赖

模拟器依赖Docker、Python3、NodeJS(如果需要启动前端)、Go(如果需要启动数据采集器)进行构建和运行

在Ubuntu下已自带Python3、NodeJS解释器
Docker可以使用以下命令进行安装

```bash

sudo apt install docker.io

```

Go可以使用snap进行安装

```bash

sudo snap install --classic go1.20

```

如果是server版本系统，请在 [All Release - The Go Programming Language](https://go.dev/dl/)找到对应的`tar.gz`包下载后进行安装

### 1.1 构建镜像

进入container-base目录下, 依次进入

1. `build_ubuntu`
2. `build_python`
3. `satellite_node`

文件夹，执行`./build_dockerfile.sh`脚本
如果提示`no such file or directory`, 请为其添加可执行权限
```bash

sudo chmod +x ./build_dockerfile.sh

```

如果需要启动监控器，请进入`sub-modules/satellite-monitor-master`文件夹

依次执行
```bash

make build
make wrap

```

### 1.2 安装项目依赖

项目的控制器和前端需要安装项目依赖

安装控制器依赖, 进入`node-manager`文件夹, 运行

```bash

sudo pip3 install -r requiremnents.txt

```

如果需要启动前端，则需要安装前端依赖，进入`ui`文件夹, 运行

```bash

npm install

```

### 1.3 启动

启动控制器需要进入控制器文件夹`node-manager`, 执行

```bash
sudo python3 main.py
```

如果需要启动前端, 则新开一个终端, 进入`ui`文件夹, 执行

```bash

npm run start
# 如果需要不同机器访问，则执行
# npm run start 0.0.0.0

```

## 2. 开发

### 2.1 开发分支约定

开发分支采用`dev -> release -> main`形式

个人开发需要新建`${Nickname}/${dev-type}`分支，在分支上进行开发, 该分支涉及内容开发完成后，提交`Merge Request`进入release分支。

定期release分支测试可以跑通后合入main分支

`dev-type`类型

1. feat: 功能开发分支，即feature
2. fix: 修复bug
3. refactor: 重构，例如重命名等
4. test: 测试样例分支，如果有的话
5. doc: 文档分支，撰写部分文档
6. chore: 其他分支，写gitignore啥的

### 2.2 commit格式约定

commit格式即commit message 格式的约定，用来表述你这个commit的类型、涉及的改动以及实现的功能，格式为

```
[${dev-type}](${涉及的文件}): ${实现的功能}
// ${dev-type}的定义与分支命名的类型相同
```

> 虽然有时候我会比较懒的写涉及的文件

例如，有一个commit, 是实现清理容器的功能，修改了main.py, 则对应的commit message即为

```
[feat](main.py): 新增清理容器功能
```

### 2.3 merge request

TBD(To Be Done)


