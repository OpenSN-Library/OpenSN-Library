# OpenSN

> 非最终版本

[Engilish Version](readme.md) 

[TOC]

## 1. 声明

## 2. 环境准备

OpenSN依赖Docker Engine实现容器的管理，在部署OpenSN之前需要确保Docker Engine已安装

Docker Engine版本为`Docker version 24.0.5`

对于Ubuntu系统，可以使用以下命令通过apt工具进安装
```bash
sudo apt install docker.io
```
也可以访问[Install Docker Engine](https://docs.docker.com/engine/install/)获取更多安装方式信息

## 3. 安装

在进行模拟前，需要准备以下物料

1. 守护进程可执行文件与配置文件
2. 依赖服务容器镜像
3. 拓扑规则控制器
4. 虚拟网络节点需要使用的镜像

### 3.1 使用预编译安装包

OpenSN提供预编译的`Linux-amd64`架构的软件包，软件包内包含以下内容

1. 守护进程可执行文件与配置文件
2. 依赖服务容器镜像
3. 拓扑规则控制器（示例）
4. 虚拟网络节点需要使用的镜像（示例）
5. OpenSN Python控制SDK
6. 自动化星座配置生成脚本

### 3.2 从源码编译

OpenSN仓库的目录结构如下
```bash
OpenSN
├── Makefile
├── TopoConfigurators # 示例拓扑规则控制器
├── container-base # 示例容器网络镜像构建目录
├── daemon # 守护进程代码子目录
├── dependencies # 依赖服务镜像构建目录
├── doc # 开发文档[TBD]
├── readme.md
├── readme_zh.md
├── tools # 自动化工具
└── ui # Web界面代码子目录
```

## 4. 启动OpenSN

### 4.1 启动前内核配置

在启动OpenSN前，为了保证虚拟网络能够正常运行，需要对以下内核参数进行配置
```bash
# 调大inotify节点数限制，容器内实时监控配置改变需要创建inotify节点
echo 4096 > /proc/sys/fs/inotify/max_user_instances
# 调大ARP表回收阈值，避免大规模虚拟网络下ARP表无法收敛
echo 8192 > /proc/sys/net/ipv4/neigh/default/gc_thresh1 
echo 16384 > /proc/sys/net/ipv4/neigh/default/gc_thresh2 
echo 32768 > /proc/sys/net/ipv4/neigh/default/gc_thresh3

# 以上参数数值可以根据单机启动容器数目改变
# 每次重启操作系统后，都需要重新配置这些参数
```

### 4.2 启动依赖服务

OpenSN将会与Etcd,InfluxDB,CodeServer三个依赖服务协同工作，各依赖服务的作用如下

* Etcd: 各机器节点、各子模块间的信息交互
* InfluxDB: 虚拟环境负载监控数据储存
* CodeServer: Topology Configurator的Web界面编辑器，非必需

### 4.3 启动守护进程

在完成4.1与4.2两个步骤后，以管理员身份运行可执行文件`NodeDaemon`即可在机器上启动OpenSN

`NodeDaemon`文件依赖同目录下的`config/config.json`文件进行配置

在用户启动`NodeDaemon`文件前，需要根据部署环境修改`config.json`文件内容

```json
{
  "app": { // NodeDaemon启动相关配置
    "is_servant": false, // 是否以从节点身份启动，设置false即以主节点身份启动
    "master_address": "172.16.208.128", // 主节点的IP地址，仅在从节点生效，用于与主节点交互
    "listen_port": 8080, // HTTP服务监听端口，仅在主节点生效
    "enable_monitor": true, // 是否启动模拟环境负载监控服务，建议启用，不然日志会有很多讨厌的Warnning
    "enable_code_server": true, // 是否启用Web代码编辑器，关闭则需要在后台终端中启动Topology Configurator
    "interface_name": "ens160", // 主网卡名，会使用该网卡第一个IPv4地址作为机器地址，从节点的master_address即为主节点机器地址
    "debug": false, // 是否启动Debug模式(会输出Debug日志)
    "instance_capacity": 400, // 机器容器节点容量，单台机器部署容器数不超过这一数字
    "monitor_interval": 1, // 负载监采样间隔,单位为秒
    "parallel_num": 64 // 并行度，最多同时执行的容器启停任务数为parallel_num, 最多同时执行的链路变更任务数为4 * parallel_num 
  },
  "dependency": { // 依赖服务相关配置，如果不了解，保持默认即可
    "etcd_port": 2379,
    "influxdb_port": 8086,
    "influxdb_token": "1145141919810",
    "influxdb_org" : "satellite_emulator",
    "influxdb_bucket": "monitor",
    "code_server_port": 8079,
    "docker_host_path": "/var/run/docker.sock"
  },
  "device": { // 预留功能，可忽略
    "FixPhysicalLink": [],
    "MultiplexPhysicalLink": []
  }
}
```

具体命令如下
```bash
sudo ./NodeDaemon
```

如果从源码编译，则`NodeDaemon`文件一般在目录`daemon/opensn-daemon/`内
如果使用预编译二进制包，则`NodeDaemon`文件在解压后文件夹的`opensn-daemon/`目录内

### 4.4 拓展OpenSN部署机器

从节点部署同样需要`内核配置`、`修改配置文件`、`启用守护进程`三个步骤

每个从节点成功与主节点建立连接后，主节点会新增两个http请求的日志输出

## 5. 运行样例程序与预期结果

在启动完成OpenSN后，可以通过地址`http://${master_address}:${listen_port}`访问OpenSN的Web控制台

在`example-project`目录下，我们提供了`示例模拟配置文件`和不同网络规模的`测试拓扑`，用户可以根据部署的硬件资源规格来选择测试拓扑

| 最小硬件规格  |      对应拓扑      |
| :-----------: | :----------------: |
|     2核4G     | example_6x11.json  |
|    8核16G     | example_31x31.json |
| 8核16G * 2台  | example_40x18.json |
| 16核32G * 4台 | example_72x22.json |

1. 将`example_emulation_configuration.json`和`example_MxN.json`上传至控制台
2. 观察控制台日志等待拓扑上传完成
3. 点击启动
4. 观察控制台日志等待启动完成
5. 在`拓扑配置器目录下`或`控制台终端`启动`Topology Configurator`
6. 在首页3D地球视图中双击节点或链路可以进入详情页面，在节点详情页面中可以启动WebShell对网络节点内部进行操作，在链路详情页面中，可以启动链路抓包功能观察链路内数据包信息。


## 6. 如何定制模拟规则

用户利用OpenSN搭建丰富的模拟环境，OpenSN在以下三个部分为用户提供了的自定义方法

1. 容器镜像
2. 拓扑控制器
3. 静态配置文件

### 6.1 定制节点类型与对应行为

OpenSN 支持任意类型的容器镜像，只需要镜像满足以下要求
1. 镜像已配置入口程序，不需要指定启动后的命令，即可以直接通过以下命令运行
```bash
sudo docker run -d ${image_name_and_version}
```
2. 镜像内根目录下`/share`目录未被占用，该目录用于OpenSN daemon挂载配置文件

满足以上条件的镜像均可以作为OpenSN模拟环境内网络节点的镜像

所以用户可以根据模拟需求，配置模拟环境内路由器、交换机、用户终端、网络服务等等各种各样的网络节点角色

如果节点需要在模拟过程中动态配置，则需要用户编写容器内的配置程序对容器内环境进行动态配置，配置所需的信息会实时更新至`/share/topo/${NODE_ID}`文件，各节点`${NODE_ID}`变量存放于容器环境变量内。

具体实现示例可以参考[卫星节点动态配置程序](container-base/satellite-router/daemon/)和[地面站节点动态配置程序](container-base/ground-station/daemon/)

### 6.2 定制模拟运行规则

容器镜像为用户提供了容器内视角的自定义方法，`Topology Configurator`则为用户提供了面向整个模拟环境的宏观视角的自定义方法。

`Topology Configurator`定义了模拟过程中的控制逻辑，决定模拟环境中的各类网络节点、链路的变化方式和变化触发条件。

同时，这一控制器也能实时地向容器内挂载目录更新配置文件，结合6.1节内的动态配置器，实现对模拟环境的全视角控制。

[Topology Configurator示例](TopoConfigurators/Standard/main.py)是一个卫星网络+地面站模拟环境的控制器示例，用户可以参考这一示例根据自己的模拟需求实现自定义的控制器。

[OpenSN SDK For Python](TopoConfigurators/opensn/)是基于Python语言的OpenSN SDK, 利用这一SDK，用户可以通过Python语言控制OpenSN模拟环境内的节点、链路、地理位置信息、容器配置等对象。

### 6.3 定制模拟规模

我们也为相同的模拟规则下，不同的星座规模和初始拓扑类型提供了自定义方式。即在第5节内用到的配置文件。

由于模拟的规模可能高达数百或数千颗卫星，用户可能需要使用自动化脚本来生成拓扑配置文件，可以参考我们在[配置文件生成脚本](tools/namespace-generator/tle_generator.py)处提供了生成脚本示例来定制自己的拓扑配置文件。

## 7. 已知问题

1. 针对较大规模的网络拓扑，受限于单一操作系统的进程调度能力，需要利用多台机器进行模拟，仅在单台机器上部署可能会降低模拟稳定性，导致路由无法收敛
2. 在早期版本在模拟结束后可能存在容器、链路残留问题，后续进行修复后虽未再发生，但是测试数量不足，如果存在容器、链路残留问题，可以执行[环境清理脚本](/daemon/scripts/delete_link_prefix.sh)清理模拟环境
