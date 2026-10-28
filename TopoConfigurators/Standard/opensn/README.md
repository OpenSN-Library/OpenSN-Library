# opensn（纯 etcd 版 SDK）

OpenSN 的 Python SDK。**只连 etcd，不调用 daemon 的任何 HTTP 接口。**

它只封装**平台协议**：键名、JSON 字段、单位、以及 daemon 原来放在 HTTP handler 里的
"分配机器 / 生成 ID / 领链路序号 / 回填连接 / 启停 / 类型配置 / WebShell 登记"等服务端逻辑。

**不包含**节点类型与镜像、容器配置格式、地址规划、拓扑变化策略——这些是用户自己的模拟设计，
示例见 `../Standard/`（`physics.py`、`addressing.py`、`container_config.py`、`main.py`）。

```python
from opensn import OpenSN

sn = OpenSN(["10.0.0.1:2379"])      # 或 OpenSN.from_env()
sn.wait_ready(timeout=30, nodes=1)

print([(n.node_index, n.ipv4, n.free_instance) for n in sn.nodes.list()])
print([i.instance_id for i in sn.instances.of_type("MyOwnType")])
```

## 安装与首次使用

```bash
pip install -r requirements.txt      # 只有 etcd3
```

首次在某个环境使用：先跑 `python3 tools/probe_etcd_api.py`，确认 etcd 客户端的事务/监听接口
与预期一致。不同版本的 python-etcd3 在元数据属性上略有差异；如果事务接口不可用，可设
`OPENSN_ETCD_COMPAT=1` 退化为逐条写（非原子，仅保证可用）。

示例里用到的 SGP4 星历计算不是 SDK 的一部分（`pip install ephem`），由用户代码自己决定。

## 平台能力

```python
# 读（本地镜像：一次快照 + watch；循环里 0 次 etcd 往返）
sn.nodes.list() / .get(k)
sn.instances.list() / .of_node(k) / .of_type("任意类型名") / .started() / .get(id)
sn.links.list() / .of(inst) / .active() / .between(a, b)
sn.positions.get(id)
sn.emulation.instance_types()

# 写（单条；在 batch/step 里自动入队合并）
inst.start() / inst.stop() / inst.remove()
inst.push_config(content)          # 内容任意：dict/list→JSON，str/bytes→原样
inst.push_config_if_absent(content)
inst.get_config() / inst.config_exists() / inst.delete_config()
inst.update(image="new-image", extra={"k": "v"}, start=True)   # 字段级修改（读最新→合并→CAS）
link.connect() / link.disconnect() / link.remove()
link.update(extra={"note": "x"}, enable=True)     # 字段级修改（跨机两份记录各自保留 node_index）
link.set(delay_s=0.023, loss_ratio=0.015, bandwidth_bps=8_000_000, connect=True)
link.set(delay_s=0.023, side=0)                   # 只写某一端（0/1）；默认两端
link.set_raw({"connect": 1, "delay": 1000})       # 自定义链路参数直接给整表
link.get_parameters(side=None) / link.remove_parameter(side=None)
sn.positions.set(inst, pos) / sn.positions.delete(inst)

# 批量提交：去重 + 跳过没变 + 按操作数分片 + 版本冲突重试
with sn.batch():
    for link in sn.links.active():
        link.set(connect=True)

# 步骤循环（读本地镜像，写先入队，进入下一轮前提交）
for step in sn.steps(interval=5):
    step.positions.set(inst, pos)

# 控制面（全部直接写 etcd；不需要 HTTP）
sn.emulation.set_instance_types({"任意类型名": {"image": "任意镜像", "container_envs": {}, "resource_limit": {"nano_cpu": "50M", "memory_byte": "128M"}}})
sn.emulation.update_instance_type("任意类型名", image="新镜像")
sn.emulation.remove_instance_type("任意类型名")
topo = sn.emulation.deploy_topology(instances=[InstanceSpec("任意类型名", extra={...})], links=[LinkSpec(0, 1)])
sn.emulation.start() / .stop() / .reset() / .status() / .wait_running(True, 60)

# WebShell（命令可自定义）
sn.shell.instance(inst)                      # 默认 docker exec 进容器 bash
sn.shell.instance(inst, command="sh", args=["-c", "ip addr"])
sn.shell.link(link)                          # 默认 tcpdump
sn.shell.request(node_index, "my_shell", "bash", [], expire_minute=5)

# 平台协议之外的自由数据（同样走批量写）
sn.kv.put("/my_state/plan", {"step": 1})
sn.kv.get("/my_state/plan") / sn.kv.get_with_revision("/my_state/plan")
sn.kv.cas_update("/my_state/plan", lambda cur: {"step": (cur or {}).get("step", 0) + 1})  # 读改写带版本校验
```

## 扩展点（把"用户策略"留在用户代码里）

| 想自定义 | 用什么 |
| --- | --- |
| 节点类型、镜像、环境变量、资源 | `emulation.set_instance_types({类型名: {image, container_envs, resource_limit}})` |
| 实例上的业务字段（你自己的轨道参数、位置、角色…） | `InstanceSpec(type, extra={...})`，平台原样存储、不做解释 |
| 容器内配置格式 | `inst.push_config(任意内容)`；格式由你的容器镜像决定 |
| 地址规划 | 自己算好后通过 `LinkSpec(address_infos=[{...}, {...}])` 或 `links.create(..., address_infos=...)` 传入 |
| 调度/放机器策略 | `deploy_topology(placement=fn(instances, nodes)->None)`，或直接指定 `InstanceSpec.node_index` |
| 拓扑变化策略（换星/断链/重连） | 自己用 `links.of()/create()/remove()` 组合，示例见 Standard 的 `switch_station` |
| 链路参数里的自定义字段 | `link.set_raw({...})`（daemon 认识哪些由 daemon 的链路类型决定） |
| 自定义键值状态 | `sn.kv`（`get/put/delete`，`cas_update` 做带版本校验的读改写）；需要镜像/订阅时用 `OpenSN(extra_prefixes=["/my_state"])` |
| 链路两端不同参数 | `link.set(..., side=0/1)` 或 `link.set_raw(table, side=0/1)` |

## 单位（SDK 对外 -> etcd 里的整数）

| 概念 | 写法 | 写进 etcd |
| --- | --- | --- |
| 时延 | `delay_s=0.023` | `23000`（微秒；daemon 下发时再 +1µs） |
| 丢包 | `loss_ratio=0.015` | `150`（万分之一） |
| 带宽 | `bandwidth_bps=8_000_000` | `1000000`（字节/秒） |
| 通断 | `connect=True/False` | `1/0` |

## 边界

- **只承诺"写进 etcd"**：daemon 不回写实际运行状态；容器/网卡是否生效只能看 daemon 日志、`docker ps`、`tc qdisc show`。
- **指标不在 SDK 范围**：监控数据在 InfluxDB，不在 etcd。
- **直写没有服务端校验**：SDK 做字段/引用/范围校验，`deploy_topology(..., dry_run=True)` 可先看将写入的键值。
- **默认行为与 daemon 的两处差异**：`reset()` 默认把 `running` 置 false；建拓扑写入顺序为实例先、链路后。

## 测试

```bash
make test     # 内存 etcd 桩：单位、schema、批量写、建拓扑、启停、自定义调度/配置/kv、watch
make check    # 与 daemon 的 Go 源码比对字段名与键名
```
