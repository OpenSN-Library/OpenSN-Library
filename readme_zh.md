# OpenSN

## 1. 声明

## 2. 环境准备

OpenSN依赖Docker Engine实现容器的管理，在部署OpenSN之前需要确保Docker Engine已安装

Docker Engine版本为`Docker version 24.0.5`

对于Ubuntu系统，可以使用以下命令通过apt工具进安装
```bash
sudo apt install docker.io
```

## 3. 安装

在进行模拟前，需要准备以下物料

1. 守护进程可执行文件与配置文件
    * 预编译二进制安装包
    * 从源代码编译
2. 依赖服务容器镜像
    * etcd
    * influxdb
    * code-server
3. 拓扑规则控制器代码（脚本）
4. 虚拟网络节点需要使用的镜像

### 3.1 守护进程可执行文件与配置文件

#### 下载预编译二进制安装包

#### 从源码编译

### 3.2 依赖服务容器镜像

### 3.3 拓扑规则控制器

### 3.4 虚拟网络节点镜像

## 4. 启动OpenSN

## 5. 运行模拟示例

## 6. 如何定制模拟规则

## 7. 已知问题