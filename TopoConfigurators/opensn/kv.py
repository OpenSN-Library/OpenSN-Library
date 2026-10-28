"""原始键值通道：给"平台协议之外"的自定义状态用。

SDK 本身只认识平台的实体键；用户如果要在 etcd 里放自己的数据（例如调度决策、
实验参数、外部工具的状态），可以用这个通道，写入同样走批量/分片/冲突重试。

自定义的键默认不会进本地镜像（SDK 不知道你要用哪些前缀）。需要镜像/订阅时，
创建客户端时传 ``extra_prefixes=["/my_state"]``。
"""

import json
from typing import Dict, List

from .errors import Conflict, NotFound
from .transport import CREATE
from .write import write_delete, write_put_bytes


class KV:
    def __init__(self, ctx):
        self._ctx = ctx

    # ---- 读（直接查 etcd，不强依赖本地镜像） ----
    def get(self, key: str, default=None):
        raw, _meta = self._ctx.transport.get(key)
        if raw is None:
            return default
        try:
            return json.loads(raw.decode("utf-8"))
        except Exception:
            return raw

    def get_raw(self, key: str):
        raw, _meta = self._ctx.transport.get(key)
        return raw

    def get_prefix(self, prefix: str) -> Dict[str, object]:
        out = {}
        for key, raw, _meta in self._ctx.transport.get_prefix(prefix):
            try:
                out[key] = json.loads(raw.decode("utf-8"))
            except Exception:
                out[key] = raw
        return out

    def require(self, key: str):
        value = self.get(key)
        if value is None:
            raise NotFound("key", key)
        return value

    def get_with_revision(self, key: str):
        """返回 (值, mod_revision)；键不存在时返回 (None, None)。"""
        raw, meta = self._ctx.transport.get(key)
        if raw is None:
            return None, None
        return _decode(raw), (meta.mod_revision if meta is not None else None)

    def cas_update(self, key: str, update_fn, tries: int = 5):
        """带版本校验的读改写：`update_fn(当前值) -> 新值`，版本没变才写入。

        值可以是 dict/list/str/bytes；键不存在时当前值为 None。冲突自动重试。
        """
        for _ in range(tries):
            raw, meta = self._ctx.transport.get(key)
            current = _decode(raw) if raw is not None else None
            new_value = update_fn(current)
            payload = _encode(new_value)
            expect = CREATE if meta is None else meta.mod_revision
            revision = self._ctx.transport.txn_puts([(key, payload, expect)])
            if revision:
                self._ctx.store.apply_local(key, payload, revision)
                return new_value
        raise Conflict(key, tries)

    # ---- 写（走批量队列；在 batch/step 里会合并提交） ----
    def put(self, key: str, value) -> None:
        if isinstance(value, bytes):
            raw = value
        elif isinstance(value, str):
            raw = value.encode("utf-8")
        else:
            raw = json.dumps(value, ensure_ascii=False).encode("utf-8")
        write_put_bytes(self._ctx, key, raw)

    def put_if_absent(self, key: str, value) -> None:
        if isinstance(value, bytes):
            raw = value
        elif isinstance(value, str):
            raw = value.encode("utf-8")
        else:
            raw = json.dumps(value, ensure_ascii=False).encode("utf-8")
        write_put_bytes(self._ctx, key, raw, CREATE)

    def delete(self, key: str) -> None:
        write_delete(self._ctx, key)

    def delete_prefix(self, prefix: str) -> None:
        for key in list(self.get_prefix(prefix).keys()):
            self.delete(key)


def _decode(raw):
    try:
        return json.loads(raw.decode("utf-8"))
    except Exception:
        return raw


def _encode(value) -> bytes:
    if isinstance(value, bytes):
        return value
    if isinstance(value, str):
        return value.encode("utf-8")
    return json.dumps(value, ensure_ascii=False).encode("utf-8")
