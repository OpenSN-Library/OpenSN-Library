"""本示例的地址规划：每条链路一对 /30 地址。

这是**用户策略**，不是平台协议：平台不分配地址，也不规定你必须用 /30、IPv4 或 10.0.0.0。
换地址方案（IPv6、按轨道编号编址、从控制器配置读）只改这里。
"""

import ipaddress
from typing import Dict, List

from opensn import OpenSN

BASE = "10.0.0.0"
PREFIX = 30


class AddressPool:
    def __init__(self, sn: OpenSN, base: str = BASE, prefix: int = PREFIX):
        self._sn = sn
        self._base = ipaddress.ip_network(f"{base}/{prefix}", strict=False)
        self._prefix = prefix
        self._reserved: List[str] = []

    def _used(self) -> Dict[str, None]:
        used = {}
        for link in self._sn.links.list():
            for addr in link.address_infos or []:
                value = (addr or {}).get("IPV4")
                if not value:
                    continue
                try:
                    used[str(ipaddress.ip_network(value, strict=False))] = None
                except ValueError:
                    continue
        return used

    def alloc(self):
        used = self._used()
        candidate = self._base
        while str(candidate) in used or str(candidate) in self._reserved:
            candidate = ipaddress.ip_network(
                f"{ipaddress.IPv4Address(int(candidate.network_address) + candidate.num_addresses)}/{self._prefix}",
                strict=False,
            )
        hosts = list(candidate.hosts())
        self._reserved.append(str(candidate))
        return f"{hosts[0]}/{self._prefix}", f"{hosts[1]}/{self._prefix}"
