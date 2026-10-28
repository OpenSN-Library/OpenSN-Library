"""SDK 入口：只连 etcd，不调用 daemon HTTP。"""

import time
from typing import Dict, Iterator, List, Optional

from . import keys
from .errors import NotReady, Unsupported
from .resources.emulation import Emulation
from .resources.instances import Instances
from .resources.links import Links
from .resources.nodes import Nodes
from .resources.positions import Positions
from .kv import KV
from .resources.shell import Shell
from .schema.entities import Instance, Link, Node, Position
from .step import Steps
from .store import Change, Store
from .transport import EtcdTransport
from .write import WriteBatch


class _Context:
    """各集合共享的连接、镜像与配置。"""

    def __init__(self, transport: EtcdTransport, store: Store, chunk_ops: int, cas_retries: int):
        self.transport = transport
        self.store = store
        self.chunk_ops = chunk_ops
        self.cas_retries = cas_retries
        self.delay_quantum_s = None
        self.nodes = None
        self.instances = None
        self.links = None
        self.positions = None
        self.emulation = None
        self.shell = None
        self.kv = None


class Snapshot:
    def __init__(self, ctx):
        self.instances: List[Instance] = ctx.instances.list()
        self.links: List[Link] = ctx.links.list()
        self.positions: Dict[str, Position] = ctx.positions.map()
        self.nodes: List[Node] = ctx.nodes.list()

    def of_type(self, type_name: str) -> List[Instance]:
        return [inst for inst in self.instances if inst.type == type_name]

    def links_of(self, inst: Instance) -> List[Link]:
        return [link for link in self.links if any(e.instance_id == inst.instance_id for e in link.end_infos)]


class OpenSN:
    def __init__(
        self,
        endpoints=None,
        client=None,
        watch: bool = True,
        poll_interval: float = 2.0,
        chunk_ops: int = 120,
        cas_retries: int = 3,
        timeout: float = 10.0,
        extra_prefixes=None,
    ):
        transport = EtcdTransport(endpoints=endpoints, client=client, timeout=timeout)
        store = Store(transport, watch=watch, poll_interval=poll_interval, extra_prefixes=extra_prefixes)
        store.snapshot()
        self._ctx = _Context(transport, store, chunk_ops=chunk_ops, cas_retries=cas_retries)
        ctx = self._ctx
        ctx.nodes = Nodes(ctx)
        ctx.instances = Instances(ctx)
        ctx.links = Links(ctx)
        ctx.positions = Positions(ctx)
        ctx.kv = KV(ctx)
        ctx.emulation = Emulation(ctx)
        ctx.shell = Shell(ctx)
        store.start()

    @classmethod
    def from_env(cls, **kwargs) -> "OpenSN":
        """读 OPENSN_ETCD_ENDPOINTS（优先）或 ETCD_ADDR / ETCD_PORT。"""
        return cls(**kwargs)

    # ---- 集合 ----
    @property
    def nodes(self) -> Nodes:
        return self._ctx.nodes

    @property
    def instances(self) -> Instances:
        return self._ctx.instances

    @property
    def links(self) -> Links:
        return self._ctx.links

    @property
    def positions(self) -> Positions:
        return self._ctx.positions

    @property
    def emulation(self) -> Emulation:
        return self._ctx.emulation

    @property
    def shell(self) -> Shell:
        return self._ctx.shell

    @property
    def kv(self) -> KV:
        """自定义键值通道（平台协议之外的数据）。"""
        return self._ctx.kv

    # ---- 就绪 ----
    def wait_ready(self, timeout: float = 30.0, nodes: int = 1) -> None:
        deadline = time.monotonic() + float(timeout)
        last_error = None
        while time.monotonic() < deadline:
            try:
                if self._ctx.transport.health():
                    self._ctx.store.refresh()
                    if len(self._ctx.nodes.list()) >= int(nodes):
                        return
            except Unsupported:
                raise
            except Exception as exc:  # 网络/etcd 未起
                last_error = exc
            time.sleep(0.3)
        raise NotReady(
            f"etcd not ready or fewer than {nodes} node(s) registered"
            + (f": {last_error}" if last_error else "")
        )

    # ---- 读快照 / 订阅 ----
    def snapshot(self, force: bool = False) -> Snapshot:
        if force:
            self._ctx.store.refresh()
        return Snapshot(self._ctx)

    def watch(self, collection=None) -> Iterator[Change]:
        prefix = None
        if collection is not None:
            name = getattr(collection, "__class__", type(collection)).__name__
            mapping = {
                "Instances": "/instances/",
                "Links": "/links/",
                "Positions": "/position/",
                "Nodes": "/nodes/",
            }
            prefix = mapping.get(name)
        while True:
            changes = self._ctx.store.drain_changes()
            matched = [c for c in changes if prefix is None or c.key.startswith(prefix)]
            for change in matched:
                yield change
            if not changes:
                time.sleep(0.2)

    # ---- 写 ----
    def batch(self) -> WriteBatch:
        return WriteBatch(self._ctx)

    def steps(self, interval: float = 5.0, delay_quantum_s: Optional[float] = None) -> Steps:
        return Steps(self._ctx, interval=interval, delay_quantum_s=delay_quantum_s)

    def cache_size(self) -> int:
        return self._ctx.store.cache_size()

    # ---- 生命周期 ----
    def close(self) -> None:
        self._ctx.store.close()
        self._ctx.transport.close()

    def __enter__(self) -> "OpenSN":
        return self

    def __exit__(self, exc_type, exc, tb) -> None:
        self.close()
