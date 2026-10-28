"""实例：读取、新建、启停、删除，以及"任意内容"的实例配置通道。"""

import json
from typing import Dict, List, Optional

from .. import keys
from ..errors import NotFound, ValidationError
from ..provisioning import create_instance, remove_instance
from ..schema.entities import Instance
from ..transport import CREATE
from ..write import write_delete, write_put, write_put_bytes


class Instances:
    def __init__(self, ctx):
        self._ctx = ctx

    # ---- 读 ----
    def _scan(self) -> List[Instance]:
        out = []
        for key, raw in self._ctx.store.items(keys.node_prefix()):
            if "/instances/" not in key:
                continue
            try:
                inst = Instance.from_json(raw)
            except Exception:
                continue
            inst._ctx = self._ctx
            out.append(inst)
        return out

    def list(self) -> List[Instance]:
        return self._scan()

    def of_node(self, node_index: int) -> List[Instance]:
        out = []
        for key, raw in self._ctx.store.items(keys.instances(node_index) + "/"):
            inst = Instance.from_json(raw)
            inst._ctx = self._ctx
            out.append(inst)
        return out

    def of_type(self, type_name: str) -> List[Instance]:
        return [inst for inst in self._scan() if inst.type == type_name]

    def started(self) -> List[Instance]:
        return [inst for inst in self._scan() if inst.start]

    def get(self, instance_id: str) -> Instance:
        matches = [inst for inst in self._scan() if inst.instance_id == instance_id]
        if not matches:
            raise NotFound("instance", instance_id)
        if len(matches) > 1:
            raise ValidationError("instance_id", instance_id, "duplicate id on multiple nodes")
        return matches[0]

    # ---- 新建 ----
    def create(
        self,
        type_name: str,
        extra=None,
        device_need=None,
        node_index: Optional[int] = None,
        charge_capacity: bool = True,
    ) -> Instance:
        return create_instance(
            self._ctx,
            type_name=type_name,
            extra=extra,
            device_need=device_need,
            node_index=node_index,
            charge_capacity=charge_capacity,
        )

    # ---- 写 ----
    def _key(self, inst: Instance) -> str:
        return keys.instance(inst.node_index, inst.instance_id)

    def _fresh(self, inst: Instance) -> Instance:
        raw = self._ctx.store.obj(self._key(inst))
        if raw is None:
            raise NotFound("instance", self._key(inst))
        fresh = Instance.from_json(raw)
        fresh._ctx = self._ctx
        return fresh

    def _save(self, inst: Instance) -> None:
        key = self._key(inst)
        expect = self._ctx.store.mod_revision(key)
        write_put(self._ctx, key, inst, expect if expect is not None else CREATE)

    def _set_start(self, inst: Instance, flag: bool) -> None:
        fresh = self._fresh(inst)
        fresh.start = bool(flag)
        inst.start = bool(flag)
        self._save(fresh)

    def _remove(self, inst: Instance) -> None:
        remove_instance(self._ctx, self._fresh(inst))

    def _update(self, inst: Instance, fields: Dict) -> None:
        editable = {"name", "image", "extra", "device_need", "resource", "start"}
        bad = set(fields) - editable
        if bad:
            raise ValidationError(
                "fields", sorted(bad), f"editable fields: {sorted(editable)}; identity/placement cannot be changed"
            )
        fresh = self._fresh(inst)
        for name, value in fields.items():
            setattr(fresh, name, value)
        self._save(fresh)
        for name, value in fields.items():
            setattr(inst, name, value)

    def _delete_config(self, inst: Instance) -> None:
        fresh = self._fresh(inst)
        write_delete(self._ctx, keys.instance_config(fresh.node_index, fresh.instance_id))

    # ---- 实例配置通道：内容由用户决定，SDK 不解释 ----
    def _push_config(self, inst: Instance, content, if_absent: bool) -> None:
        if isinstance(content, bytes):
            raw = content
        elif isinstance(content, str):
            raw = content.encode("utf-8")
        else:
            raw = json.dumps(content, ensure_ascii=False).encode("utf-8")
        key = keys.instance_config(inst.node_index, inst.instance_id)
        write_put_bytes(self._ctx, key, raw, CREATE if if_absent else None)
