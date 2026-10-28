"""链路：读取（跨机两份合并成一个对象）、新建/删除、通断、改参数。"""

import json
from typing import Dict, List, Optional, Tuple

from .. import keys
from ..errors import NotFound, ValidationError
from ..provisioning import create_link, remove_link
from ..schema.entities import Instance, Link
from ..schema.units import bandwidth_bytes, delay_us, loss_raw
from ..transport import CREATE
from ..write import write_delete, write_put

CONNECT = "connect"
DELAY = "delay"
LOSS = "loss"
BANDWIDTH = "bandwidth"


class Links:
    def __init__(self, ctx):
        self._ctx = ctx

    # ---- 读 ----
    def _records(self) -> Dict[str, List[Tuple[str, int, dict]]]:
        """link_id -> [(etcd key, node_index, raw)]，跨机链路会有两条。"""
        grouped: Dict[str, List[Tuple[str, int, dict]]] = {}
        for key, raw in self._ctx.store.items(keys.node_prefix()):
            if "/links/" not in key:
                continue
            link_id = keys.split_last(key)
            grouped.setdefault(link_id, []).append((key, int(raw.get("node_index", -1)), raw))
        return grouped

    def _merge(self, link_id: str, records: List[Tuple[str, int, dict]]) -> Link:
        records.sort(key=lambda item: item[1])
        link = Link.from_json(records[0][2])
        link._ctx = self._ctx
        link.cross_machine = len({node for _k, node, _r in records}) > 1
        return link

    def list(self) -> List[Link]:
        return [self._merge(link_id, records) for link_id, records in self._records().items()]

    def get(self, link_id: str) -> Link:
        records = self._records().get(link_id)
        if not records:
            raise NotFound("link", link_id)
        return self._merge(link_id, records)

    def of(self, inst: Instance) -> List[Link]:
        out = []
        for link in self.list():
            if any(end.instance_id == inst.instance_id for end in link.end_infos):
                out.append(link)
        return out

    def active(self) -> List[Link]:
        return [link for link in self.list() if link.enable]

    def between(self, a: Instance, b: Instance) -> Optional[Link]:
        for link in self.of(a):
            ids = {end.instance_id for end in link.end_infos}
            if b.instance_id in ids and len(ids) >= 2:
                return link
        return None

    # ---- 建链 ----
    def create(
        self,
        a: Instance,
        b: Instance,
        type: str = "vlink",
        address_infos=None,
        init_parameter=None,
        extra=None,
        enable=None,
    ) -> Link:
        """建一条链路。地址、参数、enable 都由调用方给；SDK 不替用户规划网络。"""
        return create_link(
            self._ctx,
            a,
            b,
            link_type=type,
            init_parameter=init_parameter,
            address_infos=address_infos,
            extra=extra,
            enable=enable,
        )

    def remove_between(self, a: Instance, b: Instance) -> int:
        removed = 0
        while True:
            link = self.between(a, b)
            if link is None:
                return removed
            self._remove(link)
            removed += 1

    # ---- 写 ----
    def _record_keys(self, link: Link) -> List[str]:
        out = []
        for key, raw in self._ctx.store.items(keys.node_prefix()):
            if "/links/" in key and keys.split_last(key) == link.link_id:
                out.append(key)
        return out

    def _save(self, link: Link) -> None:
        """保存链路。

        跨机链路在 etcd 里是两条记录，除了 node_index 之外内容相同。这里按 key 逐条读回
        各自的 node_index 再写，避免把合并对象的 node_index 覆盖到第二份记录上。
        """
        keys_ = self._record_keys(link)
        if not keys_:
            raise NotFound("link", link.link_id)
        for key in keys_:
            raw = self._ctx.store.obj(key) or {}
            payload = link.to_dict()
            payload["node_index"] = int(raw.get("node_index", link.node_index))
            expect = self._ctx.store.mod_revision(key)
            write_put(self._ctx, key, payload, expect if expect is not None else CREATE)

    def _update(self, link: Link, fields: Dict) -> None:
        editable = {"enable", "type", "address_infos", "extra", "parameter"}
        bad = set(fields) - editable
        if bad:
            raise ValidationError(
                "fields", sorted(bad), f"editable fields: {sorted(editable)}; topology changes use remove/create"
            )
        if "address_infos" in fields and len(fields["address_infos"] or []) != 2:
            raise ValidationError("address_infos", fields["address_infos"], "must be a 2-element array")
        fresh = self.get(link.link_id)
        for name, value in fields.items():
            setattr(fresh, name, value)
        self._save(fresh)
        for name, value in fields.items():
            setattr(link, name, value)

    def _node_indexes(self, link: Link, side=None) -> List[int]:
        if side is None:
            return sorted({end.end_node_index for end in link.end_infos})
        if side not in (0, 1):
            raise ValidationError("side", side, "must be 0 or 1 (link endpoint)")
        return [link.end_infos[side].end_node_index]

    def _parameter_table(self, link: Link, side=None) -> Dict[str, int]:
        node_index = self._node_indexes(link, side)[0]
        raw = self._ctx.store.obj(keys.link_parameter(node_index, link.link_id))
        return {k: int(v) for k, v in (raw or {}).items()}

    def _write_parameters(self, link: Link, table: Dict[str, int], side=None) -> None:
        # 整表覆盖；默认跨机链路两端各写一份；指定 side 时只写那一端。
        # 用 write_put，让 batch/step 里的调用自动入队，不会每条链路一次往返。
        for node_index in self._node_indexes(link, side):
            write_put(self._ctx, keys.link_parameter(node_index, link.link_id), table)

    def parameter(self, link: Link, side=None) -> Dict[str, int]:
        return self._parameter_table(link, side)

    def _set_parameters(self, link: Link, delay_s=None, loss_ratio=None, bandwidth_bps=None, connect=None, side=None) -> None:
        table = dict(self._parameter_table(link, side))
        if delay_s is not None:
            micro = delay_us(delay_s)
            quantum = getattr(self._ctx, "delay_quantum_s", None)
            if quantum:
                step = max(1, int(round(float(quantum) * 1_000_000)))
                micro = int(round(micro / step) * step)
            table[DELAY] = micro
        if loss_ratio is not None:
            table[LOSS] = loss_raw(loss_ratio)
        if bandwidth_bps is not None:
            table[BANDWIDTH] = bandwidth_bytes(bandwidth_bps)
        if connect is not None:
            table[CONNECT] = 1 if connect else 0
        for key, value in table.items():
            if isinstance(value, bool) or not isinstance(value, int):
                raise ValidationError("link_parameter", (key, value), "values must be integers")
        self._write_parameters(link, table, side)

    def _set_connect(self, link: Link, flag: bool, side=None) -> None:
        self._set_parameters(link, connect=flag, side=side)

    def _set_parameters_raw(self, link: Link, parameter: Dict[str, int], side=None) -> None:
        table = {}
        for key, value in (parameter or {}).items():
            if isinstance(value, bool) or not isinstance(value, int):
                raise ValidationError("link_parameter", (key, value), "values must be integers")
            table[str(key)] = int(value)
        self._write_parameters(link, table, side)

    def _remove_parameter(self, link: Link, side=None) -> None:
        for node_index in self._node_indexes(link, side):
            write_delete(self._ctx, keys.link_parameter(node_index, link.link_id))

    def _remove(self, link: Link) -> None:
        remove_link(self._ctx, link)
