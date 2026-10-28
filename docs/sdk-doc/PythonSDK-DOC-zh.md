# OpenSN Python SDK 使用手册

## 1. 安装与准备

环境要求：

- Python 3.9 及以上；
- 能访问平台的 etcd 服务（默认端口 2379）；
- 平台守护进程（daemon）已启动。

安装 SDK（只依赖 etcd 客户端，不需要 HTTP 库）：

```bash
pip install -r TopoConfigurators/opensn/requirements.txt
# 或者安装成本地包
pip install -e TopoConfigurators/opensn
```

etcd 地址通过环境变量提供，任选一种：

```bash
export OPENSN_ETCD_ENDPOINTS="10.0.0.1:2379"
# 或
export ETCD_ADDR="10.0.0.1"
export ETCD_PORT="2379"
```

可以写多个地址（逗号分隔），当前版本会使用第一个地址。

首次在一个环境使用时，建议先检查 etcd 客户端能力（事务和监听）：

```bash
cd TopoConfigurators
python3 opensn/tools/probe_etcd_api.py
```

如果事务接口不可用，可以设置 `OPENSN_ETCD_COMPAT=1` 退化为逐条写入。这种方式仍可工作，但一次批量写入不再保证整体成功。

开发或升级 SDK 时，可以检查 SDK 字段与平台代码是否一致：

```bash
cd TopoConfigurators
python3 opensn/tools/check_contract.py
```

## 2. 连接、关闭与就绪

```python
from opensn import OpenSN

sn = OpenSN(["10.0.0.1:2379"])       # 也可以 OpenSN.from_env()
sn.wait_ready(timeout=30, nodes=1)   # 等 etcd 可连接，且至少注册了 1 个节点

try:
    print([(n.node_index, n.ipv4, n.free_instance) for n in sn.nodes.list()])
finally:
    sn.close()
```

也可以使用上下文管理器：

```python
from opensn import OpenSN

with OpenSN.from_env() as sn:
    sn.wait_ready()
    ...
```

构造函数参数：

| 参数 | 默认值 | 说明 |
| --- | --- | --- |
| `endpoints` | 环境变量 | etcd 地址列表，例如 `["10.0.0.1:2379"]` |
| `client` | `None` | 直接传入 etcd 客户端对象，主要用于测试 |
| `watch` | `True` | 是否监听 etcd 变化来维护本地数据副本 |
| `poll_interval` | `2.0` | 不启用监听时，刷新本地副本的间隔（秒） |
| `chunk_ops` | `120` | 一次批量写入的最大操作数 |
| `cas_retries` | `3` | 版本冲突时的重试次数 |
| `timeout` | `10.0` | etcd 客户端超时时间（秒） |
| `extra_prefixes` | `None` | 额外需要同步到本地副本的自定义键前缀 |

`wait_ready(timeout, nodes)` 会等待 etcd 可连接，并且 `/nodes` 下的节点数量达到 `nodes`。它只证明节点记录存在，不能证明机器仍在运行。

## 3. 快速开始

下面的例子先配置一种节点类型，再创建两个实例和一条链路，然后启动模拟。

```python
from opensn import OpenSN
from opensn.schema.configs import InstanceSpec, LinkSpec

sn = OpenSN.from_env()
sn.wait_ready(timeout=30)

# 1) 配置节点类型：类型名和镜像由使用者决定
sn.emulation.set_instance_types({
    "Router": {
        "image": "my-registry/router:1.0",
        "container_envs": {"MODE": "test"},
        "resource_limit": {"nano_cpu": "200M", "memory_byte": "256M"},
    }
})

# 2) 创建两个实例和一条链路（地址由使用者给出）
topo = sn.emulation.deploy_topology(
    instances=[
        InstanceSpec("Router", extra={"role": "a"}),
        InstanceSpec("Router", extra={"role": "b"}),
    ],
    links=[
        LinkSpec(0, 1, address_infos=[{"IPV4": "10.0.0.1/30"}, {"IPV4": "10.0.0.2/30"}]),
    ],
)
print(topo.instance_ids)   # 与输入顺序一一对应
print(topo.link_ids)

# 3) 启动模拟（写入期望状态，实际启动由 daemon 完成）
sn.emulation.start()
sn.emulation.wait_running(True, timeout=60)

# 4) 运行结束后停止并清理
sn.emulation.stop()
sn.emulation.reset(clean=True)
sn.close()
```

## 4. 平台对象一览

| 对象 | 说明 | 获取方式 |
| --- | --- | --- |
| 节点 | 平台中的一台物理机 | `sn.nodes` |
| 实例 | 一个容器，可表示路由器、卫星、终端等任意角色 | `sn.instances`、`sn.links.of(inst)` |
| 链路 | 连接两个实例的网络链路 | `sn.links` |
| 链路参数 | 时延、丢包、带宽、通断 | `link.set()`、`sn.links.parameter()` |
| 位置 | 实例的经纬度和高度 | `sn.positions` |
| 实例配置 | 写给容器内 `/share/topo/<实例编号>` 文件的内容 | `inst.push_config()` |
| 模拟配置 | 实例类型到镜像、资源、环境变量的映射 | `sn.emulation` |
| 终端 | 在节点上打开容器终端或链路抓包终端 | `sn.shell` |
| 自定义键 | 平台对象之外、由使用者自己管理的数据 | `sn.kv` |

## 5. 读取数据

SDK 在启动时会读取一次全部平台数据，之后通过监听 etcd 变化维护一份本地副本。日常读取都从本地副本获取，一般不会再访问 etcd，速度很快。

### 5.1 节点

```python
nodes = sn.nodes.list()          # 全部节点
node = sn.nodes.get(0)           # 指定节点
mapping = sn.nodes.map()         # {node_index: Node}

print(node.ipv4, node.mac, node.free_instance, node.cpu_core)
```

节点由 daemon 注册，SDK 只读，不提供修改节点记录的方法。

### 5.2 实例

```python
instances = sn.instances.list()                 # 全部实例
node_instances = sn.instances.of_node(0)        # 某台机器上的实例
routers = sn.instances.of_type("Router")        # 某个类型的实例
running = sn.instances.started()                # start 为真的实例
inst = sn.instances.get("3f2a9c1d")             # 按实例编号获取

print(inst.instance_id, inst.type, inst.node_index, inst.start)
print(inst.extra)                               # 使用者自己写入的字段
```

实例编号在整个平台内唯一。如果发现同一个编号出现在多台机器上，`get` 会直接报错，不会自行选择一个。

### 5.3 链路

```python
links = sn.links.list()                 # 全部链路
link = sn.links.get("5c1e8a02")         # 按链路编号获取
linked = sn.links.of(inst)              # 与某个实例相连的链路
enabled = sn.links.active()             # enable 为真的链路
between = sn.links.between(a, b)        # 两个实例之间的链路（没有则返回 None）

print(link.addresses)                   # {"实例编号": "10.0.0.1/30", ...}
print(link.parameter)                   # 参数快照（建链时的初值）
print(link.get_parameters())            # 当前链路参数
```

跨机链路在平台内部存有两条记录。`sn.links.list()` 会把它合并成一个 `Link` 对象，使用时不需要关心这一点。

### 5.4 位置

```python
position = sn.positions.get("3f2a9c1d")
all_positions = sn.positions.map()      # {实例编号: Position}

print(position.latitude, position.longitude, position.altitude)  # 弧度、米
print(position.degrees())                                        # 转换为度
```

位置只用于展示和计算，不直接影响网络。

### 5.5 一致快照与变更通知

```python
snap = sn.snapshot()                 # 当前时刻的实例、链路、位置、节点
snap = sn.snapshot(force=True)       # 先从 etcd 重新读取一次再生成快照

print(len(snap.instances), len(snap.links))
print(snap.of_type("Router"))        # 在快照里按类型筛选
print(snap.links_of(inst))           # 在快照里查某个实例的链路
```

变更通知：

```python
for change in sn.watch(sn.links):
    print(change.kind, change.key, change.value)   # kind 为 "put" 或 "delete"
```

`watch` 是阻塞迭代，可以传入 `sn.links`、`sn.instances`、`sn.positions`、`sn.nodes` 只关注某一类对象；不传则返回所有变化。

## 6. 创建、修改、删除

### 6.1 批量部署拓扑

```python
topo = sn.emulation.deploy_topology(
    instances=[InstanceSpec("Router"), InstanceSpec("Server")],
    links=[LinkSpec(0, 1)],
    nodes=None,              # 可选：限定可用的节点编号
    placement=None,          # 可选：自定义分配实例到机器的函数
    charge_capacity=True,    # 是否扣减节点剩余容量
    dry_run=False,           # True 时只返回将要写入的数据，不写 etcd
)
```

`deploy_topology` 会依次完成：检查类型是否存在、给实例分配机器、生成实例编号和链路编号、分配链路序号、组装链路记录、写回实例与链路、维护实例之间的连接关系。

`InstanceSpec` 字段：

| 字段 | 说明 |
| --- | --- |
| `type` | 类型名，必须已经存在于模拟配置中 |
| `extra` | 自定义字符串键值对，平台原样保存 |
| `device_need` | 设备需求，例如独占网卡 |
| `node_index` | 指定机器；不填则由 SDK 分配 |

`LinkSpec` 字段：

| 字段 | 说明 |
| --- | --- |
| `a`、`b` | 两端实例在 `instances` 列表中的下标 |
| `type` | 链路类型，默认 `"vlink"` |
| `init_parameter` | 建链时的初始参数 |
| `address_infos` | 两端地址，固定两个元素 |
| `extra` | 自定义键值对 |

自定义分配策略时，传入一个函数，函数接收实例列表和节点列表，为每个实例设置 `node_index`：

```python
def place_on_nodes(instances, nodes):
    for index, inst in enumerate(instances):
        inst.node_index = nodes[index % len(nodes)].node_index

topo = sn.emulation.deploy_topology(instances, links, placement=place_on_nodes)
```

想先看写什么而不真正写入时，使用 `dry_run=True`，返回的 `topo.plan` 是“键 -> 内容”的字典。

### 6.2 单个实例与链路

```python
inst = sn.instances.create("Router", extra={"role": "edge"})
inst.remove()                       # 删除实例，同时清理它的链路、参数、配置和位置

link = sn.links.create(
    a, b,
    type="vlink",
    address_infos=[{"IPV4": "10.0.0.5/30"}, {"IPV4": "10.0.0.6/30"}],
    init_parameter={"delay": 1000},
    extra={"note": "backup"},
)
link.remove()
sn.links.remove_between(a, b)       # 删除两个实例之间的所有链路，返回删除数量
```

建链时 SDK 会自动领取链路序号、写链路记录、维护两端实例的连接关系。链路类型默认是 `vlink`；如果写入了平台不认识的类型，SDK 会给出警告，平台会忽略该链路。

### 6.3 修改字段

直接改字段适合少量修改，SDK 会先读取最新数据，再合并修改并写回：

```python
inst.update(image="my-registry/router:2.0", extra={"role": "core"}, start=True)
link.update(enable=True, extra={"note": "updated"})
```

实例可改的字段：`name`、`image`、`extra`、`device_need`、`resource`、`start`。
链路可改的字段：`enable`、`type`、`address_infos`、`extra`、`parameter`。

实例编号、所在机器、链路两端的实例、链路序号等身份或拓扑字段不能通过 `update` 修改；需要换端点时先删除再重建。

`save()` 会整条覆盖对象的所有字段，容易覆盖别人刚改的内容，一般只在明确知道自己在做什么时使用。

### 6.4 删除类操作

```python
inst.delete_config()                # 删除实例配置（容器内的配置文件也会被删除）
sn.positions.delete(inst)           # 删除单个位置
sn.positions.delete_all()           # 删除所有位置
link.remove_parameter()             # 删除链路参数（两端）
link.remove_parameter(side=1)       # 只删某一端
sn.emulation.remove_instance_type("DeprecatedType")   # 删除一种实例类型
```

### 6.5 模拟配置（实例类型）

```python
types = sn.emulation.instance_types()      # {类型名: InstanceTypeConfig}

sn.emulation.set_instance_types({          # 整体替换所有类型（未列出的类型会被删除），running 保持不变
    "Router": {"image": "router:1", "container_envs": {}, "resource_limit": {"nano_cpu": "100M", "memory_byte": "128M"}},
})

sn.emulation.update_instance_type("Router", image="router:2")   # 只改一个类型
sn.emulation.remove_instance_type("Router")                      # 删除一个类型
```

`nano_cpu` 支持 `"50M"`、`"2G"` 这类写法（按 1000 进制），`memory_byte` 支持 `"128M"`、`"1G"`（按 1024 进制）。

### 6.6 实例配置（容器内文件）

平台会把实例配置的内容原样写到容器内 `/share/topo/<实例编号>` 文件。内容格式由使用者和容器镜像约定，SDK 不解析。

```python
config = {
    "name": inst.name,
    "links": {link.link_id: link.addresses for link in sn.links.of(inst)},
}
inst.push_config(config)                    # 整体写入
inst.push_config_if_absent(config)          # 只有不存在时才写
inst.get_config()                           # 读回内容（能解析成 JSON 就返回对象）
inst.config_exists()                        # 是否存在
inst.delete_config()                        # 删除
```

注意：修改链路连接关系不会自动更新实例配置，需要由控制器代码显式调用 `push_config`。

### 6.7 位置

```python
from opensn.schema.entities import Position

sn.positions.set(inst, Position(latitude=0.1, longitude=0.2, altitude=550000.0))
sn.positions.set_all({inst_a.instance_id: pos_a, inst_b.instance_id: pos_b})
sn.positions.delete(inst)
```

经纬度用弧度，高度用米。也可以使用 `Position.from_degrees(纬度, 经度, 高度)` 创建。

### 6.8 链路参数

```python
link.set(delay_s=0.023, loss_ratio=0.015, bandwidth_bps=8_000_000, connect=True)
link.set(delay_s=0.023, side=0)          # 只写链路一端所在机器
link.set_raw({"connect": 1, "delay": 12000}, side=1)   # 直接给整张参数表
print(link.get_parameters(side=0))       # 读某一端的当前参数
link.remove_parameter(side=1)            # 删除某一端的参数
```

参数含义和写法：

| 参数 | 写法 | 说明 |
| --- | --- | --- |
| 时延 | `delay_s=0.023` | 秒；平台内部使用微秒，下发时会再加 1 微秒 |
| 丢包 | `loss_ratio=0.015` | 比例，0.015 表示 1.5% |
| 带宽 | `bandwidth_bps=8_000_000` | 比特每秒 |
| 通断 | `connect=True/False` | 是否连通 |

平台保存的是整张参数表。`set` 会先读取当前表再合并修改，未指定的参数保持不变；`set_raw` 则用传入的表整体替换。跨机链路默认两端写同一份参数，需要两端不同时用 `side=0/1`。`set_raw` 的参数表里所有值都必须是整数。

### 6.9 启动、停止与重置

```python
sn.emulation.start()        # 实例 start 置真、链路 enable 置真、模拟 running 置真
sn.emulation.stop()
sn.emulation.reset()        # 默认只删除链路和实例，并归还机器容量
sn.emulation.reset(clean=True)          # 同时清理参数、实例配置和位置
sn.emulation.reset(keep_running=True)   # 保留 running 字段，与平台旧行为一致

sn.emulation.status()                   # 当前 running 字段
sn.emulation.wait_running(True, timeout=60)
```

`start` 和 `stop` 只写入期望状态，实际启动或停止由 daemon 完成；`wait_running` 等待的是 etcd 中的 `running` 字段，不代表容器已经启动完成。

### 6.10 终端（WebShell）

```python
shell = sn.shell.instance(inst)           # 默认进入容器 bash
print(shell.url)                          # http://地址:端口/

shell = sn.shell.instance(inst, command="sh", args=["-c", "ip addr"])
shell = sn.shell.link(link)               # 默认在节点上抓包
shell = sn.shell.link(link, side=1, command="tcpdump", args=["-e", "-i", link.link_id])

shell = sn.shell.request(                 # 完全自定义命令
    node_index=0,
    webshell_id="my_shell",
    command="bash",
    args=[],
    expire_minute=10,
)
```

终端进程运行在节点机器上，不是在容器内；容器终端需要 `docker exec`。同一个终端编号重复申请会复用已有终端。

### 6.11 自定义键

平台对象之外的数据可以放在自己的键下：

```python
sn.kv.put("/my_state/plan", {"step": 1})          # 存为 JSON
sn.kv.put("/my_state/raw", "any text")            # 存为字符串
print(sn.kv.get("/my_state/plan"))
print(sn.kv.get_raw("/my_state/plan"))            # 原始字节
print(sn.kv.get_prefix("/my_state/"))             # 批量读取
sn.kv.delete("/my_state/raw")
sn.kv.delete_prefix("/my_state/")

# 读取后修改再写回，带版本校验，冲突自动重试
sn.kv.cas_update("/my_state/plan", lambda cur: {"step": (cur or {}).get("step", 0) + 1})
value, revision = sn.kv.get_with_revision("/my_state/plan")
```

如果希望自定义键也进入本地副本并出现在 `snapshot()`、`watch()` 中，创建客户端时传入前缀：

```python
sn = OpenSN(extra_prefixes=["/my_state"])
```

## 7. 批量写入与步骤循环

### 7.1 批量写入

```python
with sn.batch():
    for link in sn.links.active():
        link.set(delay_s=0.02, connect=True)
    for inst in sn.instances.started():
        sn.positions.set(inst, position_of(inst))
```

进入 `with` 后的写入先进入队列，退出时统一提交。提交时会做四件事：同一键的重复写只保留最后一次、跳过内容没有变化的写、按操作数分批、遇到版本冲突重试。

### 7.2 步骤循环

控制器通常按固定时间步运行，可以直接使用 SDK 的步骤循环：

```python
for step in sn.steps(interval=5):
    # 本轮快照，读取来自本地副本，不访问 etcd
    for inst in step.instances.started():
        step.positions.set(inst, calculate_position(inst, step.now))

    for link in step.links.active():
        link.set(connect=True)

    # 本轮写入在进入下一轮之前统一提交
```

`step.now` 是本轮开始时间，`step.instances`、`step.links` 是对象集合，`step.positions` 是位置写入入口。循环结束或异常退出时，可以调用 `steps.close()` 提交剩余写入，也可以使用 `with sn.steps(interval=5) as steps: for step in steps: ...`。

如果希望时延只按一定精度写入（例如每 1 毫秒变化才更新），可以传 `delay_quantum_s=0.001`。这会降低精度，默认关闭。

## 8. 单位与数据格式

| 内容 | SDK 写法 | 写入 etcd 的值 |
| --- | --- | --- |
| 时延 | 秒（`0.023`） | 微秒整数（`23000`） |
| 丢包 | 比例（`0.015`） | 万分之一整数（`150`） |
| 带宽 | 比特每秒（`8_000_000`） | 字节每秒（`1000000`） |
| 通断 | `True` / `False` | `1` / `0` |
| 位置 | 弧度、米 | 弧度、米 |
| 资源大小 | 字符串（`"50M"`、`"128M"`） | 整数 |

- 实例的 `extra` 是字符串到字符串的字典，平台不解释含义，适合放角色、编号、经纬度、轨道参数等业务字段。
- 实例配置的内容可以是任意 JSON、字符串或字节。
- 链路地址使用 `address_infos`，固定两个元素，每个元素是字符串到字符串的字典，常用键是 `IPV4`。

## 9. 错误处理

SDK 的异常都继承自 `OpenSNError`：

| 异常 | 含义 |
| --- | --- |
| `NotFound` | 对象或键不存在 |
| `ValidationError` | 写入前检查未通过（字段不允许修改、参数超出范围等） |
| `Conflict` | 版本冲突，重试后仍未成功 |
| `NotReady` | etcd 连接失败、等待超时、节点数量不足 |
| `EtcdError` | etcd 返回错误 |
| `Unsupported` | 当前环境不支持（例如缺少 etcd 客户端或事务接口） |

```python
import opensn

try:
    inst = sn.instances.get("not-exist")
except opensn.NotFound:
    print("实例不存在")
except opensn.ValidationError as exc:
    print("写入前检查失败：", exc)
```

## 10. 使用限制与注意事项

1. **只写 etcd，不代表已经生效。** 写完数据后，daemon 会异步执行；需要查看 daemon 日志、`docker ps`、`tc qdisc show` 等确认。
2. **只支持 etcd。** 如果部署环境只开放了 daemon 的 HTTP 端口而没有开放 etcd，SDK 无法使用。
3. **不提供监控数据。** 监控数据存放在 InfluxDB，不在 etcd 中。
4. **节点记录只读。** SDK 不提供修改节点的方法；节点剩余容量由创建和删除实例时自动维护。
5. **重置的默认行为与平台旧接口不同。** `reset()` 默认会把 `running` 置为假；需要与旧行为一致时传 `keep_running=True`。
6. **实例所在机器不能修改。** 需要把实例换到另一台机器时，只能删除后重新创建。
7. **整条保存会覆盖。** `save()` 写回对象的全部字段，可能覆盖其他程序刚改的内容；优先使用 `update()`。
8. **未知链路类型会被平台忽略。** SDK 会给出警告，但平台只实现有限的链路类型。
9. **多个程序同时修改同一对象时**，SDK 的版本校验能减少覆盖，但整条保存仍可能以最后写入为准。
10. **首次部署先运行 `probe_etcd_api.py`**，确认 etcd 客户端的事务和监听能力。

## 11. 示例

### 11.1 读取并打印平台概况

```python
from opensn import OpenSN

with OpenSN.from_env() as sn:
    sn.wait_ready()
    print("节点：")
    for node in sn.nodes.list():
        print(" ", node.node_index, node.ipv4, "剩余容量", node.free_instance)

    print("实例：")
    for inst in sn.instances.list():
        print(" ", inst.instance_id, inst.type, inst.node_index, inst.start)

    print("链路：")
    for link in sn.links.list():
        print(" ", link.link_id, link.addresses, link.get_parameters())
```

### 11.2 自定义实例分配

```python
from opensn.schema.configs import InstanceSpec

def place_by_role(instances, nodes):
    """把带 role=master 的实例放到 0 号机器，其余轮询分配。"""
    others = [node for node in nodes if node.node_index != 0] or nodes
    cursor = 0
    for inst in instances:
        if inst.extra.get("role") == "master":
            inst.node_index = 0
        else:
            inst.node_index = others[cursor % len(others)].node_index
            cursor += 1

sn.emulation.deploy_topology(
    instances=[InstanceSpec("Router", extra={"role": "master"}),
               InstanceSpec("Router", extra={"role": "worker"}),
               InstanceSpec("Router", extra={"role": "worker"})],
    links=[],
    placement=place_by_role,
)
```

### 11.3 简单控制器循环

```python
from opensn import OpenSN

sn = OpenSN.from_env()
sn.wait_ready()
sn.emulation.wait_running(True, timeout=60)

# 启动时为还没有配置的实例写一份配置
for inst in sn.instances.started():
    if not inst.config_exists():
        inst.push_config({"name": inst.name, "links": {}})

try:
    for step in sn.steps(interval=5):
        positions = {}
        for inst in step.instances.started():
            position = calculate_position(inst, step.now)   # 由使用者实现
            positions[inst.instance_id] = position
            step.positions.set(inst, position)

        for link in step.links.active():
            a, b = link.endpoints
            if a.instance_id not in positions or b.instance_id not in positions:
                continue
            link.set(
                delay_s=calculate_delay(positions[a.instance_id], positions[b.instance_id]),  # 由使用者实现
                connect=True,
                bandwidth_bps=8_000_000,
                loss_ratio=0.015,
            )
finally:
    sn.close()
```

### 11.4 自定义容器配置格式

```python
def build_config(sn, inst):
    """按自己的镜像约定生成配置，SDK 只负责原样写入。"""
    links = sn.links.of(inst)
    return {
        "hostname": inst.instance_id,
        "interfaces": {
            link.link_id: link.addresses.get(inst.instance_id)
            for link in links
        },
    }

inst.push_config(build_config(sn, inst))
```

### 11.5 用自定义键记录控制器状态

```python
sn.kv.put("/my_controller/step", {"step": 0, "last_run": ""})

for step in sn.steps(interval=5):
    ...
    sn.kv.cas_update(
        "/my_controller/step",
        lambda cur: {"step": (cur or {}).get("step", 0) + 1, "last_run": str(step.now)},
    )
```

## 附录 A：接口列表

### A.1 OpenSN

| 成员 | 说明 |
| --- | --- |
| `OpenSN(endpoints=None, client=None, watch=True, poll_interval=2.0, chunk_ops=120, cas_retries=3, timeout=10.0, extra_prefixes=None)` | 创建客户端 |
| `OpenSN.from_env(**kwargs)` | 从环境变量读取 etcd 地址 |
| `nodes` / `instances` / `links` / `positions` / `emulation` / `shell` / `kv` | 各对象集合 |
| `wait_ready(timeout=30.0, nodes=1)` | 等待 etcd 可连接且节点数量足够 |
| `snapshot(force=False)` | 返回当前快照 |
| `watch(collection=None)` | 阻塞迭代对象变化 |
| `batch()` | 进入批量写入 |
| `steps(interval=5.0, delay_quantum_s=None)` | 创建步骤循环 |
| `cache_size()` | 本地副本占用字节数 |
| `close()` | 关闭连接和后台线程 |
| `with OpenSN(...) as sn` | 自动关闭 |

### A.2 Snapshot

| 成员 | 说明 |
| --- | --- |
| `instances` | 实例列表 |
| `links` | 链路列表 |
| `positions` | `{实例编号: 位置}` |
| `nodes` | 节点列表 |
| `of_type(type_name)` | 按类型筛选实例 |
| `links_of(inst)` | 查某个实例的链路 |

### A.3 Nodes / Node

| 成员 | 说明 |
| --- | --- |
| `sn.nodes.list()` | 全部节点 |
| `sn.nodes.get(node_index)` | 指定节点 |
| `sn.nodes.map()` | `{node_index: Node}` |
| `node.node_index` | 节点编号 |
| `node.free_instance` | 剩余实例容量 |
| `node.is_master_node` | 是否主节点 |
| `node.l_3_addr_v_4` / `node.l_3_addr_v_6` / `node.l_2_addr` | 原始地址（base64 字符串） |
| `node.ipv4` / `node.ipv6` / `node.mac` | 解码后的地址 |
| `node.ns_instance_map` / `node.ns_link_map` | 平台维护的映射 |
| `node.node_link_device_info` | 设备配额 |
| `node.cpu_core` | CPU 核数 |

### A.4 Instances / Instance

| 成员 | 说明 |
| --- | --- |
| `sn.instances.list()` | 全部实例 |
| `sn.instances.of_node(node_index)` | 某台机器上的实例 |
| `sn.instances.of_type(type_name)` | 某个类型的实例 |
| `sn.instances.started()` | 已启动（`start` 为真）的实例 |
| `sn.instances.get(instance_id)` | 按编号获取 |
| `sn.instances.create(type_name, extra=None, device_need=None, node_index=None, charge_capacity=True)` | 创建实例 |
| `inst.instance_id` / `inst.name` / `inst.type` / `inst.image` | 基本信息 |
| `inst.extra` | 自定义字段 |
| `inst.device_need` | 设备需求 |
| `inst.resource` | 资源限制 |
| `inst.node_index` | 所在机器 |
| `inst.connections` | `{链路编号: 对端信息}` |
| `inst.start` | 是否期望运行 |
| `inst.start()` / `inst.stop()` | 启动/停止 |
| `inst.remove()` | 删除实例及其关联数据 |
| `inst.update(**fields)` | 修改字段 |
| `inst.save()` | 整条保存（谨慎使用） |
| `inst.push_config(content)` / `inst.push_config_if_absent(content)` | 写入实例配置 |
| `inst.get_config()` / `inst.config_exists()` / `inst.delete_config()` | 读取、检查、删除实例配置 |
| `inst.connected_to(other)` / `inst.link_to(other)` | 是否相连、到对方的链路 |

### A.5 Links / Link

| 成员 | 说明 |
| --- | --- |
| `sn.links.list()` | 全部链路 |
| `sn.links.get(link_id)` | 按编号获取 |
| `sn.links.of(inst)` | 与实例相连的链路 |
| `sn.links.active()` | 已启用的链路 |
| `sn.links.between(a, b)` | 两个实例之间的链路 |
| `sn.links.create(a, b, type="vlink", address_infos=None, init_parameter=None, extra=None, enable=None)` | 创建链路 |
| `sn.links.remove_between(a, b)` | 删除两实例间的链路 |
| `sn.links.parameter(link, side=None)` | 读取链路参数 |
| `link.link_id` / `link.type` / `link.enable` | 基本信息 |
| `link.end_infos` | 两端信息（`EndInfo` 列表，固定两个） |
| `link.address_infos` | 两端地址（固定两个） |
| `link.link_index` / `link.cross_machine` / `link.node_index` / `link.extra` | 平台字段 |
| `link.parameter` | 建链时的参数初值 |
| `link.endpoints` | `(实例, 实例)` |
| `link.addresses` | `{实例编号: 地址}` |
| `link.save()` | 整条保存 |
| `link.update(**fields)` | 修改字段 |
| `link.connect(side=None)` / `link.disconnect(side=None)` | 通断 |
| `link.remove()` | 删除链路 |
| `link.set(delay_s=None, loss_ratio=None, bandwidth_bps=None, connect=None, side=None)` | 修改参数 |
| `link.set_raw(parameter, side=None)` | 用整张表替换参数 |
| `link.get_parameters(side=None)` | 读取参数 |
| `link.remove_parameter(side=None)` | 删除参数 |

### A.6 Positions / Position

| 成员 | 说明 |
| --- | --- |
| `sn.positions.get(instance_id)` | 读取位置 |
| `sn.positions.map()` | 全部位置 |
| `sn.positions.set(inst, position)` | 写入位置 |
| `sn.positions.set_all(mapping)` | 批量写入 |
| `sn.positions.delete(inst)` | 删除位置 |
| `sn.positions.delete_all()` | 删除全部位置 |
| `Position(latitude=0.0, longitude=0.0, altitude=0.0)` | 创建位置 |
| `Position.from_degrees(latitude, longitude, altitude=0.0)` | 用度创建 |
| `position.degrees()` | 返回纬度、经度、高度的度表示 |

### A.7 Emulation / InstanceTypeConfig / TopologyResult

| 成员 | 说明 |
| --- | --- |
| `sn.emulation.instance_types()` | 读取全部实例类型 |
| `sn.emulation.set_instance_types(mapping)` | 整体替换实例类型 |
| `sn.emulation.update_instance_type(name, image=None, container_envs=None, resource_limit=None)` | 修改一个类型 |
| `sn.emulation.remove_instance_type(name)` | 删除一个类型 |
| `sn.emulation.status()` | 读取 `running` |
| `sn.emulation.wait_running(expect=True, timeout=60)` | 等待 `running` 变为指定值 |
| `sn.emulation.deploy_topology(instances, links=None, nodes=None, placement=None, charge_capacity=True, dry_run=False)` | 批量创建拓扑 |
| `sn.emulation.start()` / `stop()` | 启动/停止模拟 |
| `sn.emulation.reset(clean=False, keep_running=False)` | 重置模拟 |
| `InstanceTypeConfig.image` / `.container_envs` / `.resource_limit` | 类型配置字段 |
| `TopologyResult.instances` / `.links` / `.plan` | 创建结果 |
| `TopologyResult.instance_ids` / `.link_ids` | 与输入顺序对应的编号列表 |

### A.8 Shell / ShellInfo

| 成员 | 说明 |
| --- | --- |
| `sn.shell.request(node_index, webshell_id, command, args, expire_minute=None)` | 自定义终端 |
| `sn.shell.instance(inst, expire_minute=None, command="docker", args=None)` | 容器终端 |
| `sn.shell.link(link, expire_minute=None, side=0, command="tcpdump", args=None)` | 链路终端 |
| `ShellInfo.webshell_id` / `.addr` / `.port` / `.pid` | 终端信息 |
| `ShellInfo.url` | 终端访问地址 |

### A.9 KV

| 成员 | 说明 |
| --- | --- |
| `sn.kv.get(key, default=None)` | 读取并解析 JSON |
| `sn.kv.get_raw(key)` | 读取原始字节 |
| `sn.kv.get_prefix(prefix)` | 按前缀批量读取 |
| `sn.kv.require(key)` | 读取，不存在时报错 |
| `sn.kv.get_with_revision(key)` | 读取值和版本号 |
| `sn.kv.cas_update(key, update_fn, tries=5)` | 带版本校验的读改写 |
| `sn.kv.put(key, value)` | 写入 |
| `sn.kv.put_if_absent(key, value)` | 不存在时写入 |
| `sn.kv.delete(key)` | 删除 |
| `sn.kv.delete_prefix(prefix)` | 按前缀删除 |

### A.10 Steps / Step / StepPositions

| 成员 | 说明 |
| --- | --- |
| `sn.steps(interval=5.0, delay_quantum_s=None)` | 创建步骤循环，可迭代、可作为上下文管理器 |
| `steps.close()` | 提交剩余写入并结束 |
| `step.now` | 本轮开始时间 |
| `step.interval` | 步长（秒） |
| `step.instances` / `step.links` | 本轮对象集合 |
| `step.positions.set(inst, position)` / `step.positions.set_all(mapping)` | 本轮写入位置 |
| `step.flush()` | 立即提交本轮写入 |

### A.11 WriteBatch

| 成员 | 说明 |
| --- | --- |
| `with sn.batch() as batch` | 进入批量写入 |
| `batch.put(key, obj, expect=None)` | 写入对象或字典 |
| `batch.put_bytes(key, value, expect=None)` | 写入字节 |
| `batch.delete(key)` | 删除键 |
| `batch.size()` | 队列中的条数 |
| `batch.operation_count()` | 估算的操作数 |
| `batch.commit()` | 提交并返回操作数 |

`expect` 用于指定版本条件：不传表示直接写入；传 `-1` 表示“键必须不存在”；传具体版本号表示“版本必须一致”。

### A.12 Change

| 成员 | 说明 |
| --- | --- |
| `change.kind` | `"put"` 或 `"delete"` |
| `change.key` | 发生变化的键 |
| `change.value` | 新值（JSON 解析后的对象），删除时为 `None` |

### A.13 异常类

| 类 | 含义 |
| --- | --- |
| `OpenSNError` | 所有 SDK 异常的基类 |
| `NotFound` | 对象或键不存在 |
| `ValidationError` | 写入前检查未通过 |
| `Conflict` | 版本冲突重试后仍未成功 |
| `NotReady` | etcd 不可用、等待超时或节点不足 |
| `EtcdError` | etcd 返回错误 |
| `Unsupported` | 当前环境不支持该能力 |

## 附录 B：数据对象字段

| 对象 | 字段 |
| --- | --- |
| `InstanceSpec` | `type`、`extra`、`device_need`、`node_index` |
| `LinkSpec` | `a`、`b`、`type`、`init_parameter`、`address_infos`、`extra` |
| `EndInfo` | `instance_id`、`instance_type`、`end_node_index` |
| `Connection` | `link_id`、`instance_id`、`instance_type`、`end_node_index` |
| `DeviceRequire` | `dev_name`、`need_num`、`is_mutex` |
| `ResourceLimit` | `nano_cpu`、`memory_byte` |
| `EmulationConfig` | `running`、`type_config` |
| `InstanceTypeConfig` | `image`、`container_envs`、`resource_limit` |

## 附录 C：工具、环境变量与辅助函数

### C.1 命令行工具

| 工具 | 用途 |
| --- | --- |
| `python3 opensn/tools/probe_etcd_api.py` | 检查 etcd 客户端的事务和监听能力 |
| `python3 opensn/tools/check_contract.py` | 对比 SDK 与平台代码的字段名和键名 |
| `make test`（在 TopoConfigurators 下） | 运行 SDK 单元测试和示例冒烟测试 |
| `make check` | 运行契约检查 |

### C.2 环境变量

| 变量 | 说明 |
| --- | --- |
| `OPENSN_ETCD_ENDPOINTS` | etcd 地址，多个用逗号分隔 |
| `ETCD_ADDR` / `ETCD_PORT` | 没有上面变量时使用的 etcd 地址和端口 |
| `OPENSN_ETCD_COMPAT` | 设为 `1` 时，事务不可用则退化为逐条写入 |

### C.3 单位换算函数

```python
from opensn.schema.units import (
    parse_size_decimal, parse_size_binary,
    delay_us, loss_raw, bandwidth_bytes,
    rad_from_deg, deg_from_rad,
)
```

| 函数 | 作用 |
| --- | --- |
| `parse_size_decimal("50M")` | 按 1000 进制解析为整数 |
| `parse_size_binary("128M")` | 按 1024 进制解析为整数 |
| `delay_us(0.023)` | 秒转微秒 |
| `loss_raw(0.015)` | 比例转万分之一整数 |
| `bandwidth_bytes(8_000_000)` | 比特每秒转字节每秒 |
| `rad_from_deg(deg)` / `deg_from_rad(rad)` | 度与弧度转换 |
