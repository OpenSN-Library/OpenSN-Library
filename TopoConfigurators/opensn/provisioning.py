"""建拓扑 / 单条增删的服务端逻辑复刻。

对应 daemon 的以下行为（详见 task-spec/ref/06-etcd-only-protocols.md）：
- ``POST /api/emulation/topology``：分配机器、生成 ID、领 link_index、回填 connections
- ``DELETE /api/instance/:id``：级联删链路，并补上 daemon 漏掉的容量回补与残留清理

与 daemon 的有意差异见方案 §5.4 的"照抄与修正"表。
"""

import uuid
import warnings
from typing import Dict, Iterable, List, Optional, Tuple

from . import keys
from .errors import ValidationError as _ValidationError
from .errors import Conflict, NotFound, ValidationError
from .schema.configs import LinkSpec
from .schema.entities import (
    Connection,
    DeviceRequire,
    EmulationConfig,
    EndInfo,
    Instance,
    InstanceTypeConfig,
    Link,
    Node,
    ResourceLimit,
)
from .transport import CREATE
from .write import WriteBatch

# daemon 目前只实现了 vlink；用户可以通过 daemon 代码扩展，所以 SDK 不硬性限制类型，
# 但会提示未知类型（daemon 会静默忽略，排错很麻烦）。
KNOWN_LINK_TYPES = ("vlink", "fplink")

# link_index 会被当作跨机 VXLAN ID，24 位；daemon 不校验，SDK 在写入前挡一下。
LINK_INDEX_LIMIT = 2 ** 24


# ---------- 读配置与机器 ----------
def read_emulation_config(ctx) -> EmulationConfig:
    raw = ctx.store.obj(keys.EMULATION_CONFIG)
    if raw is None:
        return EmulationConfig(running=False, type_config={})
    return EmulationConfig.from_json(raw)


def _node_entries(ctx) -> List[Tuple[str, Node]]:
    entries = []
    for key, raw in ctx.store.items(keys.NODES + "/"):
        try:
            node = Node.from_json(raw)
        except Exception:
            continue
        entries.append((key, node))
    entries.sort(key=lambda item: item[0].encode("utf-8"))  # etcd key 字节序
    return entries


def valid_nodes(ctx, allowed: Optional[Iterable[int]] = None) -> List[Tuple[str, Node]]:
    allowed_set = set(allowed) if allowed is not None else None
    out = []
    for key, node in _node_entries(ctx):
        if node.node_index < 0 or node.cpu_core <= 0:
            continue
        if node.free_instance <= 0:
            continue
        if allowed_set is not None and node.node_index not in allowed_set:
            continue
        out.append((key, node))
    return out


def _node_valid(node: Node, inst: Instance) -> bool:
    if node.free_instance <= 0:
        return False
    devices = node.node_link_device_info or {}
    for dev, need in (inst.device_need or {}).items():
        if dev not in devices or devices[dev] < need.need_num:
            return False
    for dev, need in (inst.device_need or {}).items():
        if need.is_mutex:
            devices[dev] -= need.need_num
    return True


def place(ctx, instances: List[Instance], nodes: Optional[Iterable[int]] = None) -> List[Tuple[str, Node]]:
    """first-fit 装箱：节点按 key 字节序，选中后移到表头；返回被改动的节点记录。"""
    entries = valid_nodes(ctx, nodes)
    if not entries and instances:
        raise ValidationError("nodes", None, "no node with free capacity")
    order = list(range(len(entries)))
    touched: Dict[str, Node] = {}
    for inst in instances:
        placed = False
        for pos in list(order):
            key, node = entries[pos]
            if _node_valid(node, inst):
                inst.node_index = node.node_index
                node.free_instance -= 1
                order.remove(pos)
                order.insert(0, pos)
                touched[key] = node
                placed = True
                break
        if not placed:
            raise ValidationError("topology", inst.type, "unable to allocate instances on any node")
    return [(key, node) for key, node in touched.items()]


# ---------- 序号 ----------
def alloc_link_index_block(ctx, count: int) -> int:
    if count <= 0:
        return 0
    limit = LINK_INDEX_LIMIT
    attempts = 0
    while True:
        attempts += 1
        raw = ctx.store.raw(keys.NEXT_LINK_INDEX)
        expect = None
        if raw is None:
            start = 1
            expect = CREATE
        else:
            try:
                start = int(raw.decode("utf-8"))
            except ValueError:
                raise ValidationError("next_link_index", raw, "not an integer")
            expect = ctx.store.mod_revision(keys.NEXT_LINK_INDEX)
        if start + count > limit:
            raise ValidationError("link_index", start + count, "exceeds 24-bit VXLAN id space")
        new_value = str(start + count).encode("utf-8")
        try:
            revision = ctx.transport.txn_puts([(keys.NEXT_LINK_INDEX, new_value, expect)])
        except Exception:
            revision = 0
        if revision:
            ctx.store.apply_local(keys.NEXT_LINK_INDEX, new_value, revision)
            return start
        if attempts >= ctx.cas_retries:
            raise Conflict(keys.NEXT_LINK_INDEX, attempts)
        value, meta = ctx.transport.get(keys.NEXT_LINK_INDEX)
        ctx.store.apply_local(keys.NEXT_LINK_INDEX, value, meta.mod_revision if meta is not None else 0)


# ---------- 组装 ----------
def new_id() -> str:
    return uuid.uuid4().hex[:8]


def build_instance(spec_type: str, type_config: InstanceTypeConfig, extra, device_need, node_index: int, running: bool) -> Instance:
    instance_id = new_id()
    return Instance(
        instance_id=instance_id,
        name=f"{spec_type}_{instance_id}",
        type=spec_type,
        image=type_config.image,
        extra=dict(extra or {}),
        device_need={
            k: (v if isinstance(v, DeviceRequire) else DeviceRequire.from_json(v))
            for k, v in (device_need or {}).items()
        },
        resource=ResourceLimit(type_config.resource_limit.nano_cpu, type_config.resource_limit.memory_byte),
        node_index=node_index,
        connections={},
        start=running,
    )


def make_link_records(link_id: str, link_index: int, a: Instance, b: Instance, spec: LinkSpec, running: bool, addrs) -> List[Link]:
    end_infos = [
        EndInfo(instance_id=a.instance_id, instance_type=a.type, end_node_index=a.node_index),
        EndInfo(instance_id=b.instance_id, instance_type=b.type, end_node_index=b.node_index),
    ]
    base = Link(
        link_id=link_id,
        enable=bool(running),
        end_infos=[EndInfo(**end_infos[0].to_dict()), EndInfo(**end_infos[1].to_dict())],
        type=spec.type,
        address_infos=[dict(addrs[0]), dict(addrs[1])],
        link_index=int(link_index),
        cross_machine=a.node_index != b.node_index,
        parameter=dict(spec.init_parameter or {}),
        node_index=a.node_index,
        extra=dict(spec.extra or {}),
    )
    records = [base]
    if base.cross_machine:
        copy = Link.from_json(base.to_dict())
        copy.node_index = b.node_index
        records.append(copy)
    return records


def backfill_connections(records: List[Link], instances: List[Instance]) -> None:
    by_id = {inst.instance_id: inst for inst in instances}
    for link in records:
        a = by_id.get(link.end_infos[0].instance_id)
        b = by_id.get(link.end_infos[1].instance_id)
        if a is None or b is None:
            continue
        a.connections[link.link_id] = Connection(
            link_id=link.link_id,
            instance_id=b.instance_id,
            instance_type=b.type,
            end_node_index=b.node_index,
        )
        b.connections[link.link_id] = Connection(
            link_id=link.link_id,
            instance_id=a.instance_id,
            instance_type=a.type,
            end_node_index=a.node_index,
        )


def link_parameter_key(node_index: int, link_id: str) -> str:
    return keys.link_parameter(node_index, link_id)


# ---------- 提交 ----------
def commit_creates(ctx, instances: List[Instance], records: List[Link], nodes: List[Tuple[str, Node]], dry_run: bool = False) -> Dict[str, dict]:
    plan: Dict[str, dict] = {}
    for key, node in nodes:
        plan[key] = node.to_dict()
    for inst in instances:
        plan[keys.instance(inst.node_index, inst.instance_id)] = inst.to_dict()
    for rec in records:
        plan[keys.link(rec.node_index, rec.link_id)] = rec.to_dict()
    if dry_run:
        return plan

    batch = WriteBatch(ctx)
    for key, node in nodes:
        expect = ctx.store.mod_revision(key)
        batch.put(key, node, expect if expect is not None else CREATE)
    for inst in instances:
        batch.put(keys.instance(inst.node_index, inst.instance_id), inst, CREATE)
    for rec in records:
        batch.put(keys.link(rec.node_index, rec.link_id), rec, CREATE)
    batch.commit()

    for inst in instances:
        inst._ctx = ctx
    for rec in records:
        rec._ctx = ctx
    return plan


def _charge_nodes(ctx, instances: List[Instance]) -> List[Tuple[str, Node]]:
    """按 node_index 汇总，一次把每台机器的容量与设备配额扣掉（CAS 写）。"""
    by_node: Dict[int, int] = {}
    for inst in instances:
        by_node[inst.node_index] = by_node.get(inst.node_index, 0) + 1
    touched = []
    for node_index, count in by_node.items():
        key = keys.node(node_index)
        raw = ctx.store.obj(key)
        if raw is None:
            raise NotFound("node", key)
        node = Node.from_json(raw)
        node.free_instance = max(0, node.free_instance - count)
        touched.append((key, node))
    return touched


def _node_capacity_write(ctx, node_index: int, delta: int) -> None:
    key = keys.node(node_index)
    raw = ctx.store.obj(key)
    if raw is None:
        raise NotFound("node", key)
    node = Node.from_json(raw)
    node.free_instance = max(0, node.free_instance + delta)
    expect = ctx.store.mod_revision(key)
    batch = WriteBatch(ctx)
    batch.put(key, node, expect if expect is not None else CREATE)
    batch.commit()


# ---------- 高层入口 ----------
def deploy_topology(
    ctx,
    instances_specs,
    link_specs,
    nodes=None,
    placement=None,
    charge_capacity: bool = True,
    dry_run: bool = False,
):
    """建一批实例和链路。

    placement: 可选回调 ``fn(instances, nodes) -> None``，负责给每个 Instance 填 node_index；
               默认用与 daemon 相同的 first-fit。留空且 spec.node_index 已填时用指定机器。
    charge_capacity: 是否同时扣减 /nodes 的 free_instance。自己管理调度的用户可关掉。
    """
    emu = read_emulation_config(ctx)
    types = emu.type_config
    for spec in instances_specs:
        if spec.type not in types:
            raise ValidationError("type", spec.type, "not in /emulation_config.type_config")
    instances = [
        build_instance(spec.type, types[spec.type], spec.extra, spec.device_need, -1, emu.running)
        for spec in instances_specs
    ]
    # 显式指定机器的优先；剩下的交给 placement（默认 first-fit）
    explicit = []
    pending = []
    for inst, spec in zip(instances, instances_specs):
        if getattr(spec, "node_index", None) is not None:
            inst.node_index = int(spec.node_index)
            explicit.append(inst)
        else:
            pending.append(inst)
    nodes_touched = []
    if pending:
        if placement is not None:
            node_entries = valid_nodes(ctx, nodes)
            placement(pending, [node for _key, node in node_entries])
            for inst in pending:
                if inst.node_index < 0:
                    raise ValidationError("node_index", None, "placement did not assign a node")
            if charge_capacity:
                nodes_touched = _charge_nodes(ctx, pending)
        elif charge_capacity:
            nodes_touched = place(ctx, pending, nodes)
        else:
            node_entries = valid_nodes(ctx, nodes)
            if not node_entries:
                raise ValidationError("nodes", None, "no node with free capacity")
            order = [node for _key, node in node_entries]
            for inst in pending:
                for node in order:
                    if node.free_instance > 0:
                        inst.node_index = node.node_index
                        break
                if inst.node_index < 0:
                    raise ValidationError("nodes", None, "unable to assign instances")
    records: List[Link] = []
    if link_specs:
        if dry_run:
            raw_next = ctx.store.raw(keys.NEXT_LINK_INDEX)
            start_index = int(raw_next.decode("utf-8")) if raw_next else 1
        else:
            start_index = alloc_link_index_block(ctx, len(link_specs))
    else:
        start_index = 0
    for offset, spec in enumerate(link_specs):
        if not (0 <= spec.a < len(instances)) or not (0 <= spec.b < len(instances)):
            raise ValidationError("end_indexes", [spec.a, spec.b], "out of range")
        if not spec.type:
            raise ValidationError("type", spec.type, "link type must not be empty")
        if spec.type not in KNOWN_LINK_TYPES:
            warnings.warn(f"link type {spec.type!r} is not known to daemon; it will be ignored silently")
        a, b = instances[spec.a], instances[spec.b]
        link_id = new_id()
        addrs = list(spec.address_infos) if spec.address_infos is not None else [{}, {}]
        if len(addrs) != 2:
            raise ValidationError("address_infos", addrs, "must be a 2-element array")
        records.extend(
            make_link_records(link_id, start_index + offset, a, b, spec, emu.running, addrs)
        )
    backfill_connections(records, instances)
    plan = commit_creates(ctx, instances, records, nodes_touched, dry_run=dry_run)
    return instances, records, plan


def create_instance(
    ctx,
    type_name: str,
    extra=None,
    device_need=None,
    node_index: Optional[int] = None,
    charge_capacity: bool = True,
) -> Instance:
    emu = read_emulation_config(ctx)
    if type_name not in emu.type_config:
        raise ValidationError("type", type_name, "not in /emulation_config.type_config")
    inst = build_instance(type_name, emu.type_config[type_name], extra, device_need, -1, emu.running)
    nodes_touched = []
    if node_index is not None:
        inst.node_index = int(node_index)
        if charge_capacity:
            nodes_touched = _charge_nodes(ctx, [inst])
    else:
        nodes_touched = place(ctx, [inst], None) if charge_capacity else []
        if not charge_capacity:
            raise ValidationError("node_index", None, "node_index is required when charge_capacity=False")
    commit_creates(ctx, [inst], [], nodes_touched)
    return inst


def create_link(ctx, a: Instance, b: Instance, init_parameter=None, address_infos=None, extra=None, link_type: str = "vlink", enable: Optional[bool] = None) -> Link:
    if not link_type:
        raise ValidationError("type", link_type, "link type must not be empty")
    if link_type not in KNOWN_LINK_TYPES:
        warnings.warn(f"link type {link_type!r} is not known to daemon; it will be ignored silently")
    emu = read_emulation_config(ctx)
    if address_infos is None:
        address_infos = [{}, {}]
    start = alloc_link_index_block(ctx, 1)
    link_id = new_id()
    spec = LinkSpec(a=0, b=0, type=link_type, init_parameter=init_parameter or {}, address_infos=address_infos, extra=extra or {})
    records = make_link_records(link_id, start, a, b, spec, emu.running if enable is None else enable, address_infos)
    commit_creates(ctx, [], records, [])
    # connections 用 CAS 回填（实例已存在，可能被别的写入方改过）
    for inst, peer in ((a, b), (b, a)):
        raw = ctx.store.obj(keys.instance(inst.node_index, inst.instance_id))
        if raw is None:
            continue
        fresh = Instance.from_json(raw)
        fresh.connections[link_id] = Connection(
            link_id=link_id, instance_id=peer.instance_id, instance_type=peer.type, end_node_index=peer.node_index
        )
        expect = ctx.store.mod_revision(keys.instance(inst.node_index, inst.instance_id))
        batch = WriteBatch(ctx)
        batch.put(keys.instance(inst.node_index, inst.instance_id), fresh, expect if expect is not None else CREATE)
        batch.commit()
    # 参数必须在链路记录之后写（daemon 否则丢弃事件）
    if init_parameter:
        link = records[0]
        link._ctx = ctx
        link.set(**{
            "delay_s": init_parameter.get("delay", 0) / 1e6 if "delay" in init_parameter else None,
            "loss_ratio": init_parameter.get("loss", 0) / 10000 if "loss" in init_parameter else None,
            "bandwidth_bps": init_parameter.get("bandwidth", 0) * 8 if "bandwidth" in init_parameter else None,
            "connect": bool(init_parameter.get("connect", 1)),
        })
    return records[0]


def remove_link(ctx, link: Link) -> None:
    records = []
    for key in ctx.store.keys(keys.node_prefix()):
        if key.endswith(f"/links/{link.link_id}"):
            records.append((key, ctx.store.obj(key) or {}))
    if not records:
        raise NotFound("link", link.link_id)
    node_indexes = {int(raw.get("node_index", -1)) for _key, raw in records}
    batch = WriteBatch(ctx)
    for key, _raw in records:
        batch.delete(key)
    for node_index in node_indexes:
        batch.delete(link_parameter_key(node_index, link.link_id))
    for end in link.end_infos:
        inst_key = keys.instance(end.end_node_index, end.instance_id)
        raw = ctx.store.obj(inst_key)
        if raw is None:
            continue
        fresh = Instance.from_json(raw)
        if link.link_id in fresh.connections:
            del fresh.connections[link.link_id]
            expect = ctx.store.mod_revision(inst_key)
            batch.put(inst_key, fresh, expect if expect is not None else CREATE)
    batch.commit()


def remove_instance(ctx, inst: Instance) -> None:
    key = keys.instance(inst.node_index, inst.instance_id)
    raw = ctx.store.obj(key)
    if raw is None:
        raise NotFound("instance", key)
    fresh = Instance.from_json(raw)
    batch = WriteBatch(ctx)
    for link_id in list(fresh.connections.keys()):
        for node_key in ctx.store.keys(keys.node_prefix()):
            if node_key.endswith(f"/links/{link_id}"):
                batch.delete(node_key)
        for node_key in ctx.store.keys(keys.node_prefix()):
            if node_key.endswith(f"/link_parameter/{link_id}"):
                batch.delete(node_key)
        for peer in (fresh.connections[link_id],):
            peer_key = keys.instance(peer.end_node_index, peer.instance_id)
            peer_raw = ctx.store.obj(peer_key)
            if peer_raw is None:
                continue
            peer_inst = Instance.from_json(peer_raw)
            if link_id in peer_inst.connections:
                del peer_inst.connections[link_id]
                expect = ctx.store.mod_revision(peer_key)
                batch.put(peer_key, peer_inst, expect if expect is not None else CREATE)
    batch.delete(key)
    batch.delete(keys.instance_config(inst.node_index, inst.instance_id))
    batch.delete(keys.position(inst.instance_id))
    batch.commit()
    _node_capacity_write(ctx, inst.node_index, +1)


def reset(ctx, clean: bool = False, keep_running: bool = False) -> None:
    emu = read_emulation_config(ctx)
    node_counts: Dict[int, int] = {}
    batch = WriteBatch(ctx)
    for key in list(ctx.store.keys(keys.node_prefix())):
        if "/links/" in key:
            raw = ctx.store.obj(key) or {}
            batch.delete(key)
            if clean and raw.get("link_id"):
                batch.delete(link_parameter_key(int(raw.get("node_index", -1)), raw["link_id"]))
        elif "/instances/" in key:
            raw = ctx.store.obj(key) or {}
            node_index = int(raw.get("node_index", -1))
            node_counts[node_index] = node_counts.get(node_index, 0) + 1
            batch.delete(key)
            if raw.get("instance_id"):
                if clean:
                    batch.delete(keys.instance_config(node_index, raw["instance_id"]))
                    batch.delete(keys.position(raw["instance_id"]))
    batch.commit()
    for node_index, count in node_counts.items():
        node_key = keys.node(node_index)
        raw = ctx.store.obj(node_key)
        if raw is None:
            continue
        node = Node.from_json(raw)
        node.free_instance = node.free_instance + count
        expect = ctx.store.mod_revision(node_key)
        b = WriteBatch(ctx)
        b.put(node_key, node, expect if expect is not None else CREATE)
        b.commit()
    if not keep_running:
        config_key = keys.EMULATION_CONFIG
        raw = ctx.store.obj(config_key)
        emu = EmulationConfig.from_json(raw or {})
        emu.running = False
        expect = ctx.store.mod_revision(config_key)
        b = WriteBatch(ctx)
        b.put(config_key, emu, expect if expect is not None else CREATE)
        b.commit()
