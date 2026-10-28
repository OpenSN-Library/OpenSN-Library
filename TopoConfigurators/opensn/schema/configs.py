"""建拓扑时的输入描述。

这里只有"要建什么"的数据结构，不含任何地址规划、容器配置格式或调度策略——
那些是用户代码的事（见 Standard/ 里的 addressing.py、container_config.py、main.py）。
"""

from dataclasses import dataclass, field
from typing import Dict, List, Optional


@dataclass
class InstanceSpec:
    """一个待创建的实例。

    type / extra / device_need 都是原样透传给平台的字段：type 对应
    /emulation_config.type_config 里的名字，extra 是用户自由定义的一组字符串
    （例如你自己的轨道参数、位置、角色标签），平台不理解它们的含义。

    node_index 留空表示"由 SDK 的默认调度（first-fit）分配"；填了就用指定的机器。
    """

    type: str
    extra: Dict[str, str] = field(default_factory=dict)
    device_need: Dict[str, dict] = field(default_factory=dict)
    node_index: Optional[int] = None


@dataclass
class LinkSpec:
    """一条待创建的链路；a/b 是 instances 列表里的下标。"""

    a: int
    b: int
    type: str = "vlink"
    init_parameter: Dict[str, int] = field(default_factory=dict)
    address_infos: Optional[List[dict]] = None
    extra: Dict[str, str] = field(default_factory=dict)
