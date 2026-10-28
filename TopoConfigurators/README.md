# TopoConfigurators：优化后的 SDK 与示例控制器

这是 [task-spec/sdk-refactor-plan.md](../task-spec/sdk-refactor-plan.md) 的落地示例。

设计边界：**SDK 只封装平台协议；节点类型/镜像、容器配置格式、地址规划、拓扑变化策略都是用户代码。**
这些策略在 `Standard/` 里实现，可以直接替换成你自己的。

```
TopoConfigurators/
├── opensn/                 # SDK（canonical）：只连 etcd，只做平台协议
│   ├── client.py           # OpenSN：连接、生命周期、batch/step
│   ├── transport.py        # etcd 客户端封装（事务、比较、watch）
│   ├── keys.py / schema/   # 键与 JSON 契约、单位换算
│   ├── resources/          # nodes / instances / links / positions / emulation / shell
│   ├── provisioning.py     # 建拓扑、单条增删（复刻 daemon HTTP handler 的服务端逻辑）
│   ├── store.py / write.py # 本地镜像 / 批量写（去重、跳过、分片、版本重试）
│   ├── kv.py               # 平台协议之外的自由键值（含 CAS 读改写）
│   ├── step.py             # 步骤循环
│   ├── tools/              # check_contract.py、probe_etcd_api.py
│   └── tests/              # 内存 etcd 桩上的单元测试
├── Standard/               # 示例控制器：所有"用户策略"都在这里
│   ├── main.py             # 调度/换星/断链策略
│   ├── physics.py          # TLE→位置、距离、时延、最近卫星、极区规则
│   ├── addressing.py       # /30 地址规划（用户策略）
│   ├── container_config.py # 容器内配置格式（用户与镜像的约定）
│   ├── config.py / instance_types.py
│   └── opensn/             # 由 ./sync_sdk.sh 从 ../opensn 同步（Standard 按原约定从自己目录 import）
├── tests/                  # 端到端冒烟测试（内存 etcd + ephem 桩）
└── sync_sdk.sh             # 同步 SDK 副本
```

## 跑测试

```bash
cd TopoConfigurators
make test      # SDK 单元测试 + Standard 冒烟测试
make check     # 契约检查（与 daemon Go 源码比对字段/键）
```

## 跑示例控制器

需要真实 etcd 和已经启动的模拟环境（拓扑上传、点启动之后）：

```bash
cd Standard
pip install -r requirements.txt
OPENSN_ETCD_ENDPOINTS=10.0.0.1:2379 python3 main.py
```

SDK 不支持 HTTP，所以不需要 daemon 的 8080 端口。

## 怎么换成自己的实验

1. 节点类型与镜像：改 `emulation.set_instance_types(...)` 的内容（或在控制台上传）。
2. 实例字段：`InstanceSpec(type, extra={...})`，`extra` 里放你自己的业务键。
3. 容器配置：改 `container_config.py` 的 `build()`，返回你的镜像认识的任何结构。
4. 地址：改 `addressing.py`，或直接在 `LinkSpec/links.create` 里给 `address_infos`。
5. 拓扑变化：改 `main.py` 里的策略函数（换星、断链、恢复），可以完全重写。
6. 自定义状态：用 `sn.kv`，需要镜像/订阅就用 `OpenSN(extra_prefixes=[...])`。

## 修改 SDK 时注意

`opensn/` 是唯一源；`Standard/opensn/` 是它的一份副本（因为 Standard 按原约定从自己目录 import）。
改完 canonical 后跑一次 `./sync_sdk.sh`。
