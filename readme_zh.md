# OpenSN

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

### 4.1 系统预配置

### 4.2 启动依赖服务

### 4.3 启动守护进程

## 5. 运行样例程序与预期结果

## 6. 如何定制模拟规则

### 6.1 定制节点类型与对应行为

### 6.2 定制模拟运行规则

## 7. 已知问题
