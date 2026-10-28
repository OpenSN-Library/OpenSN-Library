"""etcd 键构造。全 SDK 只有这里拼键字符串。

键模板与 daemon/share/key/data_keys.go 一一对应，改 daemon 时同步改这里，
并跑 tools/check_contract.py。
"""

NODES = "/nodes"
NODE_LIST = "/nodes"
EMULATION_CONFIG = "/emulation_config"
NEXT_LINK_INDEX = "/next_link_index"
POSITION = "/position"


def node(node_index: int) -> str:
    return f"/nodes/{int(node_index)}"


def node_prefix() -> str:
    return "/node_"


def instances(node_index: int) -> str:
    return f"/node_{int(node_index)}/instances"


def instance(node_index: int, instance_id: str) -> str:
    return f"{instances(node_index)}/{instance_id}"


def links(node_index: int) -> str:
    return f"/node_{int(node_index)}/links"


def link(node_index: int, link_id: str) -> str:
    return f"{links(node_index)}/{link_id}"


def link_parameters(node_index: int) -> str:
    return f"/node_{int(node_index)}/link_parameter"


def link_parameter(node_index: int, link_id: str) -> str:
    return f"{link_parameters(node_index)}/{link_id}"


def instance_config(node_index: int, instance_id: str) -> str:
    return f"/instance_config/node_{int(node_index)}/{instance_id}"


def instance_config_prefix() -> str:
    return "/instance_config/"


def position(instance_id: str) -> str:
    return f"{POSITION}/{instance_id}"


def webshell_requests(node_index: int) -> str:
    return f"/node_{int(node_index)}/webshell_request"


def webshell_request(node_index: int, webshell_id: str) -> str:
    return f"{webshell_requests(node_index)}/{webshell_id}"


def webshell_infos(node_index: int) -> str:
    return f"/node_{int(node_index)}/webshell_info"


def webshell_info(node_index: int, webshell_id: str) -> str:
    return f"{webshell_infos(node_index)}/{webshell_id}"


def split_last(key: str) -> str:
    """取键最后一段（实例 ID / 链路 ID / 节点号）。"""
    return key.rstrip("/").rsplit("/", 1)[-1]


SNAPSHOT_PREFIXES = (
    NODES,
    POSITION,
    EMULATION_CONFIG,
    "/node_",
    "/instance_config/",
)
