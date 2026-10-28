# Standard 控制器（纯 etcd 版示例）

一个卫星网络 + 地面站的模拟控制器。运行时每 `STEP_SECOND`（默认 5 秒）一步：

1. 计算所有已启动实例的位置，批量写到 `/position`（给三维展示与其他配置器看）；
2. 地面站始终连最近的卫星：换星时由 SDK 断旧链、复用旧地址、建新链、重新下发两端容器配置；
3. 跨轨道面的链路经过极区（纬度 > 66.5°）时断开；其余链路按两端距离更新时延。

## 依赖

```bash
pip install -r requirements.txt     # etcd3 + ephem
```

SDK 就在同目录的 `opensn/` 下（与 `TopoConfigurators/opensn` 是同一份，由
`TopoConfigurators/sync_sdk.sh` 同步）。

## 运行

先准备好模拟环境（上传拓扑、启动），然后：

```bash
OPENSN_ETCD_ENDPOINTS=10.0.0.1:2379 python3 main.py
# 或 ETCD_ADDR=10.0.0.1 ETCD_PORT=2379 python3 main.py
```

可选环境变量：`STEP_SECOND`、`POLAR_THRESHOLD_DEG`、`BANDWIDTH_BPS`、`LOSS_RATIO`。

## 和旧版 Standard 的差异

| | 旧版 | 这一版 |
| --- | --- | --- |
| 连接方式 | 先 HTTP 探活/取 etcd 地址 | 直接给 etcd 端点（环境变量），不发任何 HTTP 请求 |
| 地址分配 | 自己维护 `/30` 分配器 | 仍在示例里（`addressing.py`），SDK 只负责把地址写进链路记录 |
| 容器配置 | 手工拼 JSON、手动同步 2~3 次 | 仍在示例里（`container_config.py`），但由策略代码在连接变化时显式下发 |
| 状态轮询 | 每步全量读 etcd + `sleep(5)` | SDK 本地镜像（watch），循环内 0 次读 |
| 主循环 | 207 行，约七成是平台协议 | 平台协议归 SDK，策略代码约 100 行（含换星/配置格式/地址规划） |

本目录里的东西都是"用户策略"，不是平台协议：

- `physics.py`：TLE → 位置、距离、时延、最近卫星、极区判定（换成自己的轨道模型即可）
- `addressing.py`：每条链路一对 /30 地址（换成 IPv6、按轨道编址等都可以）
- `container_config.py`：容器内配置文件的结构（换成你的镜像认识的结构即可）
- `main.py`：调度与拓扑变化策略（选星、换星、断链、参数取值）

## 注意

- etcd 里只有"期望状态"，daemon 不回写实际运行状态。本控制器只保证"已经写入"，
  容器/网卡是否真的生效请查 daemon 日志、`docker ps`、`tc qdisc show`。
- 指标在 InfluxDB 里，不在本 SDK 的范围。
