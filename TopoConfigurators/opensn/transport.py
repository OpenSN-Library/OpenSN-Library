"""etcd 传输层。

只封装 etcd，不引入任何 HTTP 客户端。真实客户端用 ``etcd3``（python-etcd3），
为避免把它的细节散到各处，这一层提供：

- ``get`` / ``get_prefix`` / ``put`` / ``delete`` / ``delete_prefix``
- ``txn_puts``：一批写入（可选版本比较，见 :data:`CREATE`）
- ``txn_deletes``：一批删除
- ``watch_prefix``：返回 (事件迭代器, cancel 函数)

若是首次在某个环境运行，先跑 ``tools/probe_etcd_api.py`` 确认客户端的事务/监听接口
符合预期（不同版本的 python-etcd3 在元数据属性上略有差异）。
"""

import os
from dataclasses import dataclass
from typing import Any, Dict, Iterable, List, Optional, Tuple

from .errors import Conflict, EtcdError, NotReady, Unsupported

# CAS 的两种比较方式：CREATE 表示"键必须不存在"（create_revision == 0），
# 大于等于 0 的整数表示"mod_revision 必须等于这个值"。
CREATE = -1


@dataclass
class Meta:
    key: str = ""
    mod_revision: int = 0
    create_revision: int = 0
    version: int = 0


def _as_bytes(value: Any) -> bytes:
    if value is None:
        return b""
    if isinstance(value, bytes):
        return value
    return str(value).encode("utf-8")


def _meta_key(meta: Any) -> str:
    key = getattr(meta, "key", b"")
    if isinstance(key, bytes):
        return key.decode("utf-8", "replace")
    return str(key or "")


def _meta_from(meta: Any) -> Meta:
    if meta is None:
        return Meta()
    return Meta(
        key=_meta_key(meta),
        mod_revision=int(getattr(meta, "mod_revision", 0) or 0),
        create_revision=int(getattr(meta, "create_revision", 0) or 0),
        version=int(getattr(meta, "version", 0) or 0),
    )


def parse_endpoints(endpoints) -> List[Tuple[str, int]]:
    if not endpoints:
        env = os.environ.get("OPENSN_ETCD_ENDPOINTS", "")
        if not env:
            host = os.environ.get("ETCD_ADDR", "127.0.0.1")
            port = os.environ.get("ETCD_PORT", "2379")
            endpoints = [f"{host}:{port}"]
        else:
            endpoints = [e.strip() for e in env.split(",") if e.strip()]
    if isinstance(endpoints, str):
        endpoints = [endpoints]
    out = []
    for item in endpoints:
        item = item.strip()
        if "://" in item:
            item = item.split("://", 1)[1]
        if ":" in item:
            host, port = item.rsplit(":", 1)
            out.append((host, int(port)))
        else:
            out.append((item, 2379))
    if not out:
        raise NotReady("no etcd endpoint configured")
    return out


class EtcdTransport:
    def __init__(self, endpoints=None, client=None, timeout: float = 10.0):
        self._endpoints = parse_endpoints(endpoints) if client is None else []
        self._client = client
        self._timeout = timeout
        self._compat = os.environ.get("OPENSN_ETCD_COMPAT", "") in ("1", "true", "yes")

    # ---- 连接 ----
    @property
    def client(self):
        if self._client is None:
            try:
                import etcd3  # type: ignore
            except Exception as exc:  # pragma: no cover - 依赖缺失时给出明确提示
                raise Unsupported(
                    "etcd3 is not installed; run `pip install -r requirements.txt` "
                    f"(import error: {exc})"
                )
            host, port = self._endpoints[0]
            self._client = etcd3.client(host=host, port=port, timeout=self._timeout)
        return self._client

    def health(self) -> bool:
        try:
            self.get("/nodes")
            return True
        except Exception:
            return False

    def close(self) -> None:
        client = self._client
        if client is not None and hasattr(client, "close"):
            try:
                client.close()
            except Exception:
                pass
        self._client = None

    # ---- 基础读写 ----
    def get(self, key: str) -> Tuple[Optional[bytes], Optional[Meta]]:
        try:
            value, meta = self.client.get(key)
        except Exception as exc:
            raise EtcdError(str(exc))
        if value is None:
            return None, None
        return _as_bytes(value), _meta_from(meta)

    def get_prefix(self, prefix: str) -> List[Tuple[str, bytes, Meta]]:
        items: List[Tuple[str, bytes, Meta]] = []
        try:
            for value, meta in self.client.get_prefix(prefix):
                m = _meta_from(meta)
                items.append((m.key or _meta_key(meta), _as_bytes(value), m))
        except Exception as exc:
            raise EtcdError(str(exc))
        return items

    def put(self, key: str, value: bytes) -> Optional[Meta]:
        try:
            resp = self.client.put(key, value)
        except Exception as exc:
            raise EtcdError(str(exc))
        return self._resp_meta(key, resp)

    def delete(self, key: str) -> None:
        try:
            self.client.delete(key)
        except Exception as exc:
            raise EtcdError(str(exc))

    def delete_prefix(self, prefix: str) -> None:
        try:
            self.client.delete_prefix(prefix)
        except Exception as exc:
            raise EtcdError(str(exc))

    @staticmethod
    def _resp_meta(key: str, resp: Any) -> Optional[Meta]:
        header = getattr(resp, "header", None)
        if header is not None:
            return Meta(key=key, mod_revision=int(getattr(header, "revision", 0) or 0))
        return None

    # ---- 批量事务 ----
    def txn_puts(self, ops: Iterable[Tuple[str, bytes, Optional[int]]]) -> int:
        """一次事务写入多条，成功返回写入后的 revision，失败返回 0。

        每条可带 expect：None / CREATE / mod_revision。
        """
        ops = list(ops)
        if not ops:
            return 0
        if self._compat:
            return self._compat_puts(ops)
        client = self.client
        try:
            txns = client.transactions
            compare = []
            success = []
            for key, value, expect in ops:
                if expect is not None:
                    if expect == CREATE:
                        compare.append(txns.create(key) == 0)
                    else:
                        compare.append(txns.mod(key) == int(expect))
                success.append(txns.put(key, value))
            resp = client.transaction(compare=compare, success=success, failure=[])
            if not bool(getattr(resp, "succeeded", False)):
                return 0
            header = getattr(resp, "header", None)
            revision = int(getattr(header, "revision", 0) or 0) if header is not None else 0
            return revision or 1
        except (AttributeError, TypeError) as exc:
            raise Unsupported(
                "this etcd3 client does not support transactions as expected; "
                f"run tools/probe_etcd_api.py (error: {exc})"
            )

    def txn_deletes(self, keys: Iterable[str]) -> bool:
        keys = list(keys)
        if not keys:
            return True
        if self._compat:
            for key in keys:
                self.delete(key)
            return True
        client = self.client
        try:
            txns = client.transactions
            resp = client.transaction(compare=[], success=[txns.delete(k) for k in keys], failure=[])
            return bool(getattr(resp, "succeeded", False))
        except (AttributeError, TypeError) as exc:
            raise Unsupported(
                "this etcd3 client does not support transaction deletes as expected; "
                f"run tools/probe_etcd_api.py (error: {exc})"
            )

    def _compat_puts(self, ops) -> int:
        """兼容模式：逐条比较再写（不是原子的，只用于客户端缺事务接口时）。"""
        for key, value, expect in ops:
            if expect is not None:
                _, meta = self.get(key)
                current = None if meta is None else meta.mod_revision
                if expect == CREATE:
                    if meta is not None:
                        return 0
                elif current != expect:
                    return 0
            self.put(key, value)
        return 1

    # ---- 监听 ----
    def watch_prefix(self, prefix: str):
        client = self.client
        if not hasattr(client, "watch_prefix"):
            raise Unsupported("etcd3 client has no watch_prefix; set watch=False or upgrade etcd3")
        result = client.watch_prefix(prefix)
        if isinstance(result, tuple):
            return Watcher(*result)
        return result


class Watcher:
    """把 python-etcd3 的 (iterator, cancel) 适配成统一接口。"""

    def __init__(self, iterator, cancel):
        self._iterator = iterator
        self._cancel = cancel

    def __iter__(self):
        return self._iterator

    def close(self) -> None:
        if self._cancel is not None:
            try:
                self._cancel()
            except Exception:
                pass
