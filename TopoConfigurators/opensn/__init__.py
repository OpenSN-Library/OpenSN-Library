"""OpenSN Python SDK（纯 etcd 版）。

只通过 etcd 读写实体，不调用 daemon 的 HTTP 接口。对外入口见 :class:`opensn.client.OpenSN`。
"""

from .client import OpenSN
from .errors import (
    Conflict,
    EtcdError,
    NotFound,
    NotReady,
    OpenSNError,
    Unsupported,
    ValidationError,
)
from .schema.entities import (
    Connection,
    DeviceRequire,
    EmulationConfig,
    EndInfo,
    Instance,
    InstanceTypeConfig,
    Link,
    Node,
    Position,
    ResourceLimit,
    ShellInfo,
)
from .step import Step, Steps

__all__ = [
    "OpenSN",
    "OpenSNError",
    "NotFound",
    "ValidationError",
    "Conflict",
    "NotReady",
    "EtcdError",
    "Unsupported",
    "Node",
    "Instance",
    "Link",
    "EndInfo",
    "Connection",
    "DeviceRequire",
    "ResourceLimit",
    "Position",
    "EmulationConfig",
    "InstanceTypeConfig",
    "ShellInfo",
    "Step",
    "Steps",
]

__version__ = "0.1.0"
