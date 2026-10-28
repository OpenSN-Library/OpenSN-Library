"""模拟配置：类型配置、建拓扑、启停、重置、状态。全部直接读写 etcd。"""

import time
from dataclasses import dataclass, field
from typing import Dict, List, Optional

from .. import keys
from ..errors import NotReady, NotFound, ValidationError
from ..provisioning import alloc_link_index_block, deploy_topology, read_emulation_config, reset as do_reset
from ..schema.configs import InstanceSpec, LinkSpec
from ..schema.entities import (
    EmulationConfig,
    Instance,
    InstanceTypeConfig,
    Link,
    ResourceLimit,
)
from ..schema.units import parse_size_binary, parse_size_decimal
from ..transport import CREATE
from ..write import WriteBatch


@dataclass
class TopologyResult:
    instances: List[Instance] = field(default_factory=list)
    links: List[Link] = field(default_factory=list)
    plan: Dict[str, dict] = field(default_factory=dict)

    @property
    def instance_ids(self) -> List[str]:
        return [inst.instance_id for inst in self.instances]

    @property
    def link_ids(self) -> List[str]:
        seen = []
        for link in self.links:
            if link.link_id not in seen:
                seen.append(link.link_id)
        return seen


def _parse_type_config(name: str, data: dict) -> InstanceTypeConfig:
    if "image" not in data:
        raise ValidationError("image", name, "missing image")
    limit = data.get("resource_limit") or {}
    return InstanceTypeConfig(
        image=str(data["image"]),
        container_envs={k: str(v) for k, v in (data.get("container_envs") or {}).items()},
        resource_limit=ResourceLimit(
            nano_cpu=parse_size_decimal(limit.get("nano_cpu", 0)),
            memory_byte=parse_size_binary(limit.get("memory_byte", 0)),
        ),
    )


class Emulation:
    def __init__(self, ctx):
        self._ctx = ctx

    # ---- 类型配置 ----
    def _read(self) -> EmulationConfig:
        return read_emulation_config(self._ctx)

    def _write(self, config: EmulationConfig) -> None:
        key = keys.EMULATION_CONFIG
        expect = self._ctx.store.mod_revision(key)
        batch = WriteBatch(self._ctx)
        batch.put(key, config, expect if expect is not None else CREATE)
        batch.commit()

    def instance_types(self) -> Dict[str, InstanceTypeConfig]:
        return dict(self._read().type_config)

    def set_instance_types(self, mapping: Dict[str, dict]) -> None:
        """整体替换 type_config；running 保持原值（daemon 的 HTTP 接口也是这个语义）。"""
        current = self._read()
        parsed = {name: _parse_type_config(name, data or {}) for name, data in mapping.items()}
        self._write(EmulationConfig(running=current.running, type_config=parsed))

    def update_instance_type(
        self,
        name: str,
        image: Optional[str] = None,
        container_envs: Optional[Dict[str, str]] = None,
        resource_limit: Optional[dict] = None,
    ) -> None:
        current = self._read()
        existing = current.type_config.get(name, InstanceTypeConfig())
        if image is not None:
            existing.image = image
        if container_envs is not None:
            existing.container_envs = dict(container_envs)
        if resource_limit is not None:
            existing.resource_limit = _parse_type_config(
                name, {"image": existing.image or "x", "resource_limit": resource_limit}
            ).resource_limit
        current.type_config[name] = existing
        self._write(current)

    def remove_instance_type(self, name: str) -> None:
        current = self._read()
        if name not in current.type_config:
            raise NotFound("instance_type", name)
        del current.type_config[name]
        self._write(current)

    def status(self) -> bool:
        return self._read().running

    def wait_running(self, expect: bool = True, timeout: float = 60.0) -> bool:
        deadline = time.monotonic() + float(timeout)
        while time.monotonic() < deadline:
            if self.status() == bool(expect):
                return True
            time.sleep(0.2)
        raise NotReady(f"/emulation_config.running did not become {expect} in {timeout}s")

    # ---- 建拓扑 ----
    def deploy_topology(
        self,
        instances: List[InstanceSpec],
        links: Optional[List[LinkSpec]] = None,
        nodes: Optional[List[int]] = None,
        placement=None,
        charge_capacity: bool = True,
        dry_run: bool = False,
    ) -> TopologyResult:
        """建一批实例与链路。

        placement: 自定义调度回调 ``fn(instances, nodes)->None``；默认与 daemon 相同的 first-fit。
        charge_capacity: 是否扣 /nodes.free_instance；自己管理调度时可关掉。
        """
        inst_objs, link_objs, plan = deploy_topology(
            self._ctx,
            instances,
            links or [],
            nodes=nodes,
            placement=placement,
            charge_capacity=charge_capacity,
            dry_run=dry_run,
        )
        return TopologyResult(instances=inst_objs, links=link_objs, plan=plan)

    # ---- 启停 ----
    def _iter_raw(self, suffix: str):
        for key, raw in self._ctx.store.items(keys.node_prefix()):
            if suffix in key:
                yield key, raw

    def start(self) -> None:
        if not self.status():
            self._switch_instances(True)
            self._switch_links(True)
            self._write(EmulationConfig(running=True, type_config=self._read().type_config))
        else:
            self._switch_instances(True)
            self._switch_links(True)

    def stop(self) -> None:
        self._switch_links(False)
        self._switch_instances(False)
        if self.status():
            self._write(EmulationConfig(running=False, type_config=self._read().type_config))

    def _switch_instances(self, flag: bool) -> None:
        batch = WriteBatch(self._ctx)
        for key, raw in self._iter_raw("/instances/"):
            inst = Instance.from_json(raw)
            if inst.start == flag:
                continue
            inst.start = flag
            batch.put(key, inst, self._ctx.store.mod_revision(key))
        batch.commit()

    def _switch_links(self, flag: bool) -> None:
        batch = WriteBatch(self._ctx)
        for key, raw in self._iter_raw("/links/"):
            link = Link.from_json(raw)
            if link.enable == flag and not flag:
                continue
            if flag:
                # daemon 启用链路时会用记录里的 parameter 重设内核参数，
                # 所以要把当前参数表写回记录（只重写 /link_parameter 不会生效）。
                raw_params = self._ctx.store.obj(keys.link_parameter(link.node_index, link.link_id))
                link.parameter = {k: int(v) for k, v in (raw_params or link.parameter).items()}
            link.enable = flag
            batch.put(key, link, self._ctx.store.mod_revision(key))
        batch.commit()

    # ---- 重置 ----
    def reset(self, clean: bool = False, keep_running: bool = False) -> None:
        do_reset(self._ctx, clean=clean, keep_running=keep_running)
