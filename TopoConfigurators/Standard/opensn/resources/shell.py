"""WebShell：通过 etcd 的请求/响应键申请终端（不调用 daemon HTTP）。

协议（daemon/pkg/module/webshell_module.go）：
- 申请方写 /node_<k>/webshell_request/<id>：{webshell_id, command, args, writeable, expire_minute}
- 该节点 daemon 起 gotty 后写 /node_<k>/webshell_info/<id>：{webshell_id, addr, port, pid}
- 进程退出（超时/被杀）时 daemon 删除两个键

注意：同一个 ID 重复写 request 会被忽略；要重开必须先删 request。命令不能被改写。
"""

import json
import time
from typing import List, Optional

from .. import keys
from ..errors import NotFound, NotReady, ValidationError
from ..schema.entities import ShellInfo
from ..transport import CREATE
from ..write import write_delete, write_put

WAIT_TIMEOUT = 60.0


class Shell:
    def __init__(self, ctx):
        self._ctx = ctx

    def _wait_info(self, node_index: int, webshell_id: str, timeout: float) -> ShellInfo:
        """等待 daemon 写响应键。

        这是 daemon 侧异步动作，本地镜像不保证立刻能看到（尤其是没开 watch 时），
        所以这里直接查 etcd；一次申请只多几次读，不影响循环性能。
        """
        info_key = keys.webshell_info(node_index, webshell_id)
        req_key = keys.webshell_request(node_index, webshell_id)
        deadline = time.monotonic() + float(timeout)
        while time.monotonic() < deadline:
            raw, meta = self._ctx.transport.get(info_key)
            if raw is not None:
                self._ctx.store.apply_local(info_key, raw, meta.mod_revision if meta else 0)
                return ShellInfo.from_json(json.loads(raw.decode("utf-8")))
            req_raw, _meta = self._ctx.transport.get(req_key)
            if req_raw is None:
                raise NotReady(
                    f"webshell request for {webshell_id} disappeared before info appeared; "
                    "check daemon logs and gotty in depend-bin"
                )
            time.sleep(0.2)
        raise NotReady(f"timeout waiting for {info_key}")

    def _request(self, node_index: int, webshell_id: str, command: str, args: List[str], expire_minute: Optional[int]) -> ShellInfo:
        info_key = keys.webshell_info(node_index, webshell_id)
        existing = self._ctx.store.obj(info_key)
        if existing is not None:
            return ShellInfo.from_json(existing)
        req_key = keys.webshell_request(node_index, webshell_id)
        if self._ctx.store.obj(req_key) is not None:
            # 残留请求（上次没起成功）：先删再写，否则 daemon 认为 ID 没变会忽略
            write_delete(self._ctx, req_key)
        payload = {
            "webshell_id": webshell_id,
            "command": command,
            "args": list(args),
            "writeable": True,
            "expire_minute": int(expire_minute or 0),
        }
        write_put(self._ctx, req_key, payload, CREATE)
        return self._wait_info(node_index, webshell_id, WAIT_TIMEOUT)

    def request(
        self,
        node_index: int,
        webshell_id: str,
        command: str,
        args: List[str],
        expire_minute: Optional[int] = None,
    ) -> ShellInfo:
        """通用申请：daemon 会在该节点上以 ``command args...`` 启动 gotty。

        命令在节点宿主机上执行（不是容器内），所以容器终端需要 ``docker exec ...``。
        """
        return self._request(node_index, webshell_id, command, args, expire_minute)

    def instance(
        self,
        inst,
        expire_minute: Optional[int] = None,
        command: str = "docker",
        args: Optional[List[str]] = None,
    ) -> ShellInfo:
        instance_id = getattr(inst, "instance_id", inst)
        fresh = self._ctx.instances.get(instance_id)
        node_index = getattr(inst, "node_index", None) or fresh.node_index
        webshell_id = f"instance_{instance_id}"
        if args is None:
            # 默认进容器 bash；容器命名规则是平台的 <type>_<instance_id>
            args = ["exec", "-it", f"{fresh.type}_{instance_id}", "/bin/bash"]
        return self._request(node_index, webshell_id, command, args, expire_minute)

    def link(
        self,
        link,
        expire_minute: Optional[int] = None,
        side: int = 0,
        command: str = "tcpdump",
        args: Optional[List[str]] = None,
    ) -> ShellInfo:
        if side not in (0, 1):
            raise ValidationError("side", side, "must be 0 or 1")
        node_index = link.end_infos[side].end_node_index
        webshell_id = f"link_{link.link_id}"
        if args is None:
            args = ["-e", "-i", link.link_id]
        return self._request(node_index, webshell_id, command, args, expire_minute)
