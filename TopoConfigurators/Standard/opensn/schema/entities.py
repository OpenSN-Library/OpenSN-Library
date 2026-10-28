"""与 daemon 的 JSON 结构一一对应的数据类。

字段名严格使用 daemon/model/*.go 的 json tag。特别注意：
- 实例的设备需求字段叫 ``device_need``（不是 device_info）
- 实例类型的环境变量字段叫 ``container_envs``（不是 envs）
- 链路的 ``end_infos``/``address_infos`` 是定长 2 的数组
- 节点的 ``l_3_addr_v_4`` 等是 base64 字符串（Go []byte 的 JSON 编码）
- ``connection.end_node_index`` 是整数
"""

import base64
import json
import socket
from dataclasses import dataclass, field, fields
from typing import Any, Dict, List, Optional

from ..errors import ValidationError


def _config_key(instance) -> str:
    from ..keys import instance_config

    return instance_config(instance.node_index, instance.instance_id)


def _encode_value(value: Any) -> Any:
    if hasattr(value, "to_dict"):
        return value.to_dict()
    if isinstance(value, list):
        return [_encode_value(v) for v in value]
    if isinstance(value, dict):
        return {k: _encode_value(v) for k, v in value.items()}
    return value


def _public_fields(obj) -> Dict[str, Any]:
    return {f.name: getattr(obj, f.name) for f in fields(obj) if not f.name.startswith("_")}


class _JsonMixin:
    def to_dict(self) -> Dict[str, Any]:
        return {k: _encode_value(v) for k, v in _public_fields(self).items()}

    def to_json(self) -> str:
        return json.dumps(self.to_dict())


@dataclass
class ResourceLimit(_JsonMixin):
    nano_cpu: int = 0
    memory_byte: int = 0

    @classmethod
    def from_json(cls, data: Optional[dict]) -> "ResourceLimit":
        data = data or {}
        return cls(nano_cpu=int(data.get("nano_cpu", 0) or 0), memory_byte=int(data.get("memory_byte", 0) or 0))


@dataclass
class DeviceRequire(_JsonMixin):
    dev_name: str = ""
    need_num: int = 0
    is_mutex: bool = False

    @classmethod
    def from_json(cls, data: dict) -> "DeviceRequire":
        return cls(
            dev_name=data.get("dev_name", ""),
            need_num=int(data.get("need_num", 0) or 0),
            is_mutex=bool(data.get("is_mutex", False)),
        )


@dataclass
class Connection(_JsonMixin):
    link_id: str = ""
    instance_id: str = ""
    instance_type: str = ""
    end_node_index: int = -1

    @classmethod
    def from_json(cls, data: dict) -> "Connection":
        return cls(
            link_id=data.get("link_id", ""),
            instance_id=data.get("instance_id", ""),
            instance_type=data.get("instance_type", ""),
            end_node_index=int(data.get("end_node_index", -1)),
        )


@dataclass
class Node(_JsonMixin):
    node_index: int = -1
    free_instance: int = 0
    is_master_node: bool = False
    l_3_addr_v_4: str = ""
    l_3_addr_v_6: str = ""
    l_2_addr: str = ""
    ns_instance_map: Dict[str, str] = field(default_factory=dict)
    ns_link_map: Dict[str, str] = field(default_factory=dict)
    node_link_device_info: Dict[str, int] = field(default_factory=dict)
    cpu_core: int = 0

    @classmethod
    def from_json(cls, data: Optional[dict]) -> "Node":
        data = data or {}
        return cls(
            node_index=int(data.get("node_index", -1)),
            free_instance=int(data.get("free_instance", 0) or 0),
            is_master_node=bool(data.get("is_master_node", False)),
            l_3_addr_v_4=str(data.get("l_3_addr_v_4") or ""),
            l_3_addr_v_6=str(data.get("l_3_addr_v_6") or ""),
            l_2_addr=str(data.get("l_2_addr") or ""),
            ns_instance_map=dict(data.get("ns_instance_map") or {}),
            ns_link_map=dict(data.get("ns_link_map") or {}),
            node_link_device_info={k: int(v) for k, v in (data.get("node_link_device_info") or {}).items()},
            cpu_core=int(data.get("cpu_core", 0) or 0),
        )

    # --- 地址字段是 base64 字节，这里提供解码后的可读形式 ---
    @property
    def ipv4(self) -> str:
        return _decode_ip(self.l_3_addr_v_4, 4)

    @property
    def ipv6(self) -> str:
        return _decode_ip(self.l_3_addr_v_6, 16)

    @property
    def mac(self) -> str:
        raw = _b64decode(self.l_2_addr)
        if not raw or len(raw) < 6:
            return ""
        return ":".join(f"{b:02x}" for b in raw[-6:])


def _b64decode(text: str) -> bytes:
    if not text:
        return b""
    try:
        return base64.b64decode(text)
    except Exception:
        return b""


def _decode_ip(text: str, length: int) -> str:
    raw = _b64decode(text)
    if len(raw) != length:
        return ""
    if length == 4:
        return socket.inet_ntop(socket.AF_INET, raw)
    return socket.inet_ntop(socket.AF_INET6, raw)


@dataclass
class Instance(_JsonMixin):
    instance_id: str = ""
    name: str = ""
    type: str = ""
    image: str = ""
    extra: Dict[str, str] = field(default_factory=dict)
    device_need: Dict[str, DeviceRequire] = field(default_factory=dict)
    resource: ResourceLimit = field(default_factory=ResourceLimit)
    node_index: int = -1
    connections: Dict[str, Connection] = field(default_factory=dict)
    start: bool = False
    _ctx: Any = field(default=None, init=False, repr=False, compare=False)

    @classmethod
    def from_json(cls, data: Optional[dict]) -> "Instance":
        data = data or {}
        raw_devices = data.get("device_need") or data.get("device_info") or {}
        raw_conns = data.get("connections") or {}
        return cls(
            instance_id=str(data.get("instance_id", "") or ""),
            name=str(data.get("name", "") or ""),
            type=str(data.get("type", "") or ""),
            image=str(data.get("image", "") or ""),
            extra={k: str(v) for k, v in (data.get("extra") or {}).items()},
            device_need={k: DeviceRequire.from_json(v or {}) for k, v in raw_devices.items()},
            resource=ResourceLimit.from_json(data.get("resource")),
            node_index=int(data.get("node_index", -1)),
            connections={k: Connection.from_json(v or {}) for k, v in raw_conns.items()},
            start=bool(data.get("start", False)),
        )

    # --- 便捷只读属性 ---
    def connected_to(self, other) -> bool:
        if self.connections:
            for conn in self.connections.values():
                if conn.instance_id == other.instance_id:
                    return True
        return False

    def link_to(self, other):
        if maybe_ctx := self._ctx:
            return maybe_ctx.links.between(self, other)
        return None

    # --- 写操作（需要 _ctx；由集合或 step 注入） ---
    def _require_ctx(self):
        if self._ctx is None:
            raise ValidationError("instance", self.instance_id, "entity is detached from a client")
        return self._ctx

    def save(self):
        """整条覆盖写（不推荐，容易覆盖别人的修改）。"""
        return self._require_ctx().instances._save(self)

    def start(self):
        return self._require_ctx().instances._set_start(self, True)

    def stop(self):
        return self._require_ctx().instances._set_start(self, False)

    def remove(self):
        return self._require_ctx().instances._remove(self)

    def update(self, **fields):
        """字段级修改：先读最新记录、合并字段、再带版本号写回。

        可改：name、image、extra、device_need、resource、start。
        身份（instance_id）与所在机器（node_index）不能改。
        """
        return self._require_ctx().instances._update(self, fields)

    def delete_config(self):
        """删除实例配置（容器内 /share/topo 文件也会被 daemon 删掉）。"""
        return self._require_ctx().instances._delete_config(self)

    def push_config(self, content):
        """下发实例配置（内容任意：dict/list → JSON，str/bytes → 原样）。"""
        return self._require_ctx().instances._push_config(self, content, if_absent=False)

    def push_config_if_absent(self, content):
        return self._require_ctx().instances._push_config(self, content, if_absent=True)

    def get_config(self):
        """读回实例配置：能解析成 JSON 就返回对象，否则返回字符串；没有则 None。"""
        import json as _json

        ctx = self._require_ctx()
        raw = ctx.store.raw(_config_key(self))
        if raw is None:
            return None
        try:
            return _json.loads(raw.decode("utf-8"))
        except Exception:
            return raw.decode("utf-8", "replace")

    def config_exists(self) -> bool:
        return self._require_ctx().store.raw(_config_key(self)) is not None


@dataclass
class EndInfo(_JsonMixin):
    instance_id: str = ""
    instance_type: str = ""
    end_node_index: int = -1

    @classmethod
    def from_json(cls, data: Optional[dict]) -> "EndInfo":
        data = data or {}
        return cls(
            instance_id=str(data.get("instance_id", "") or ""),
            instance_type=str(data.get("instance_type", "") or ""),
            end_node_index=int(data.get("end_node_index", -1)),
        )


@dataclass
class Link(_JsonMixin):
    link_id: str = ""
    enable: bool = False
    end_infos: List[EndInfo] = field(default_factory=lambda: [EndInfo(), EndInfo()])
    type: str = "vlink"
    address_infos: List[Dict[str, str]] = field(default_factory=lambda: [{}, {}])
    link_index: int = 0
    cross_machine: bool = False
    parameter: Dict[str, int] = field(default_factory=dict)
    node_index: int = -1
    extra: Dict[str, str] = field(default_factory=dict)
    _ctx: Any = field(default=None, init=False, repr=False, compare=False)

    @classmethod
    def from_json(cls, data: Optional[dict]) -> "Link":
        data = data or {}
        ends = data.get("end_infos") or [{}, {}]
        addrs = data.get("address_infos") or [{}, {}]
        ends = (list(ends) + [{}, {}])[:2]
        addrs = (list(addrs) + [{}, {}])[:2]
        return cls(
            link_id=str(data.get("link_id", "") or ""),
            enable=bool(data.get("enable", False)),
            end_infos=[EndInfo.from_json(e) for e in ends],
            type=str(data.get("type", "") or ""),
            address_infos=[dict(a or {}) for a in addrs],
            link_index=int(data.get("link_index", 0) or 0),
            cross_machine=bool(data.get("cross_machine", False)),
            parameter={k: int(v) for k, v in (data.get("parameter") or {}).items()},
            node_index=int(data.get("node_index", -1)),
            extra={k: str(v) for k, v in (data.get("extra") or {}).items()},
        )

    @property
    def endpoints(self):
        if self._ctx is None:
            return tuple(self.end_infos)
        return (
            self._ctx.instances.get(self.end_infos[0].instance_id),
            self._ctx.instances.get(self.end_infos[1].instance_id),
        )

    @property
    def addresses(self) -> Dict[str, str]:
        """{'<instance_id>': 'a.b.c.d/30', ...}，取 address_infos 里的 IPV4。"""
        out = {}
        for end, addr in zip(self.end_infos, self.address_infos):
            if addr and "IPV4" in addr:
                out[end.instance_id] = addr["IPV4"]
        return out

    def _require_ctx(self):
        if self._ctx is None:
            raise ValidationError("link", self.link_id, "entity is detached from a client")
        return self._ctx

    def save(self):
        return self._require_ctx().links._save(self)

    def update(self, **fields):
        """字段级修改：可改 enable、type、address_infos、extra、parameter。

        拓扑相关字段（link_id、node_index、link_index、cross_machine、end_infos）不允许直接改，
        需要换端点的请 remove() 后重新 create()。
        """
        return self._require_ctx().links._update(self, fields)

    def connect(self, side=None):
        return self._require_ctx().links._set_connect(self, True, side=side)

    def disconnect(self, side=None):
        return self._require_ctx().links._set_connect(self, False, side=side)

    def remove(self):
        return self._require_ctx().links._remove(self)

    def set(self, delay_s=None, loss_ratio=None, bandwidth_bps=None, connect=None, side=None):
        """改链路参数。默认两端都写；side=0/1 只写指定端点所在机器。"""
        return self._require_ctx().links._set_parameters(
            self, delay_s=delay_s, loss_ratio=loss_ratio, bandwidth_bps=bandwidth_bps, connect=connect, side=side
        )

    def set_raw(self, parameter: Dict[str, int], side=None):
        """直接提交整张参数表：给自定义链路类型或平台未来新增的参数用。"""
        return self._require_ctx().links._set_parameters_raw(self, parameter, side=side)

    def get_parameters(self, side=None) -> Dict[str, int]:
        return self._require_ctx().links.parameter(self, side=side)

    def remove_parameter(self, side=None) -> None:
        """删除链路参数表（side=0/1 只删一端；默认删两端）。"""
        return self._require_ctx().links._remove_parameter(self, side=side)


@dataclass
class Position(_JsonMixin):
    latitude: float = 0.0
    longitude: float = 0.0
    altitude: float = 0.0

    @classmethod
    def from_json(cls, data: Optional[dict]) -> "Position":
        data = data or {}
        return cls(
            latitude=float(data.get("latitude", 0.0) or 0.0),
            longitude=float(data.get("longitude", 0.0) or 0.0),
            altitude=float(data.get("altitude", 0.0) or 0.0),
        )

    @classmethod
    def from_degrees(cls, latitude_deg: float, longitude_deg: float, altitude_m: float = 0.0) -> "Position":
        import math

        return cls(
            latitude=float(latitude_deg) / 180.0 * math.pi,
            longitude=float(longitude_deg) / 180.0 * math.pi,
            altitude=float(altitude_m),
        )

    def degrees(self):
        import math

        return (
            self.latitude * 180.0 / math.pi,
            self.longitude * 180.0 / math.pi,
            self.altitude,
        )


@dataclass
class InstanceTypeConfig(_JsonMixin):
    image: str = ""
    container_envs: Dict[str, str] = field(default_factory=dict)
    resource_limit: ResourceLimit = field(default_factory=ResourceLimit)

    @classmethod
    def from_json(cls, data: Optional[dict]) -> "InstanceTypeConfig":
        data = data or {}
        return cls(
            image=str(data.get("image", "") or ""),
            container_envs={k: str(v) for k, v in (data.get("container_envs") or {}).items()},
            resource_limit=ResourceLimit.from_json(data.get("resource_limit")),
        )


@dataclass
class EmulationConfig(_JsonMixin):
    running: bool = False
    type_config: Dict[str, InstanceTypeConfig] = field(default_factory=dict)

    @classmethod
    def from_json(cls, data: Optional[dict]) -> "EmulationConfig":
        data = data or {}
        return cls(
            running=bool(data.get("running", False)),
            type_config={k: InstanceTypeConfig.from_json(v) for k, v in (data.get("type_config") or {}).items()},
        )


@dataclass
class ShellInfo(_JsonMixin):
    webshell_id: str = ""
    addr: str = ""
    port: int = 0
    pid: int = 0

    @classmethod
    def from_json(cls, data: Optional[dict]) -> "ShellInfo":
        data = data or {}
        return cls(
            webshell_id=str(data.get("webshell_id", "") or ""),
            addr=str(data.get("addr", "") or ""),
            port=int(data.get("port", 0) or 0),
            pid=int(data.get("pid", 0) or 0),
        )

    @property
    def url(self) -> str:
        return f"http://{self.addr}:{self.port}/"
