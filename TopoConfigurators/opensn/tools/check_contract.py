#!/usr/bin/env python3
"""契约检查：把 SDK 的字段名/键名与 daemon 的 Go 源码对齐。

用法（在 TopoConfigurators 目录下）::

    python3 opensn/tools/check_contract.py [--daemon /path/to/daemon]

任何字段或键不一致都返回非 0，适合放进 CI。
"""

import argparse
import re
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[2]))

from opensn import keys  # noqa: E402
from opensn.schema.entities import (  # noqa: E402
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

# SDK 数据类 -> (Go 文件, Go 结构体名)
STRUCT_MAP = [
    (Node, "model/node.go", "Node"),
    (Instance, "model/instance.go", "Instance"),
    (Connection, "model/instance.go", "ConnectionInfo"),
    (DeviceRequire, "model/instance.go", "DeviceRequireInfo"),
    (Link, "model/link.go", "LinkBase"),
    (EndInfo, "model/link.go", "EndInfoType"),
    (ResourceLimit, "model/emulation_config.go", "ResourceLimit"),
    (InstanceTypeConfig, "model/emulation_config.go", "InstanceTypeConfig"),
    (EmulationConfig, "model/emulation_config.go", "EmulationInfo"),
    (Position, "model/position.go", "Position"),
    (ShellInfo, "model/webshell.go", "WebShellAllocInfo"),
]

# SDK 键 -> (Go 常量名, 生成的示例键)
KEY_MAP = [
    ("/nodes", "NodeIndexListKey"),
    ("/emulation_config", "EmulationConfigKey"),
    ("/next_link_index", "NextLinkIndexKey"),
    ("/position", "InstancePositionKey"),
    ("/node_0/instances", "NodeInstanceListKeyTemplate"),
    ("/node_0/links", "NodeLinkListKeyTemplate"),
    ("/node_0/link_parameter", "NodeLinkParameterKeyTemplate"),
    ("/instance_config/node_0", "NodeInstanceConfigKeyTemplate"),
    ("/node_0/webshell_request", "NodeWebshellRequestKeyTemplate"),
    ("/node_0/webshell_info", "NodeWebshellInfoKeyTemplate"),
]


def find_repo_root(start: Path):
    """从当前文件向上找含 daemon/model/node.go 的仓库根目录（不假设固定层级）。"""
    for candidate in [start, *start.parents]:
        if (candidate / "daemon" / "model" / "node.go").exists():
            return candidate
    return None


def parse_struct(path: Path, struct_name: str):
    text = path.read_text(encoding="utf-8", errors="replace")
    match = re.search(r"type\s+%s\s+struct\s*\{(.*?)\n\}" % re.escape(struct_name), text, re.S)
    if not match:
        return None
    fields = set()
    for line in match.group(1).splitlines():
        tag = re.search(r'json:"([^",]+)', line)
        if tag and tag.group(1) != "-":
            fields.add(tag.group(1))
    return fields


def parse_keys(path: Path):
    text = path.read_text(encoding="utf-8", errors="replace")
    found = {}
    for match in re.finditer(r"(\w+)\s*=\s*\"([^\"]+)\"", text):
        found[match.group(1)] = match.group(2)
    return found


def normalize_template(template: str) -> str:
    return re.sub(r"%d", "0", template)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--daemon", default=None, help="daemon/ 目录")
    args = parser.parse_args()
    repo = find_repo_root(Path(__file__).resolve().parent)
    daemon = Path(args.daemon) if args.daemon else (repo / "daemon" if repo else None)
    if daemon is None or not (daemon / "model" / "node.go").exists():
        print(f"skip: daemon sources not found at {daemon}")
        return 0

    problems = []
    for cls, rel, struct in STRUCT_MAP:
        go_fields = parse_struct(daemon / rel, struct)
        if go_fields is None:
            problems.append(f"{rel}: struct {struct} not found")
            continue
        sdk_fields = {f.name for f in __import__("dataclasses").fields(cls) if not f.name.startswith("_")}
        # SDK 允许"读兼容"字段（例如 device_info 别名的读取由 from_json 处理），但写出集合必须逐字一致
        missing = go_fields - sdk_fields
        extra = sdk_fields - go_fields
        if missing:
            problems.append(f"{cls.__name__}: missing daemon fields {sorted(missing)}")
        if extra:
            problems.append(f"{cls.__name__}: fields not in daemon {sorted(extra)}")

    go_keys = parse_keys(daemon / "share" / "key" / "data_keys.go")
    for sample_key, const in KEY_MAP:
        if const not in go_keys:
            problems.append(f"data_keys.go: constant {const} not found")
            continue
        expected = normalize_template(go_keys[const])
        if expected != sample_key:
            problems.append(f"key {const}: daemon={expected!r} sdk={sample_key!r}")

    if problems:
        print("CONTRACT MISMATCH")
        for item in problems:
            print(" -", item)
        return 1
    print(f"contract OK: {len(STRUCT_MAP)} structs, {len(KEY_MAP)} keys")
    return 0


if __name__ == "__main__":
    sys.exit(main())
