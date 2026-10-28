"""写队列：去重、跳过没变的值、按操作数分片、版本冲突重试。

片大小按"操作数"算：一个 compare 也算一个操作。
- 不比较的值写（位置、参数整表）：1 个操作
- 带比较的读改写（实例、链路、节点、connections、新建键）：2 个操作
"""

import contextvars
import json
from typing import Any, Dict, Iterable, List, Optional, Tuple

from .errors import Conflict, ValidationError
from .transport import CREATE, EtcdError

_current_batch: "contextvars.ContextVar[Optional[WriteBatch]]" = contextvars.ContextVar(
    "opensn_current_batch", default=None
)


def current_batch() -> Optional["WriteBatch"]:
    return _current_batch.get()


def detach_batch():
    return _current_batch.set(None)


def reset_batch(token) -> None:
    _current_batch.reset(token)


def attach_batch(batch: "WriteBatch"):
    return _current_batch.set(batch)


def encode(obj: Any) -> bytes:
    if hasattr(obj, "to_dict"):
        payload = obj.to_dict()
    elif isinstance(obj, dict):
        payload = obj
    else:
        raise ValidationError("write", obj, "expect dict or entity object")
    return json.dumps(payload, ensure_ascii=False).encode("utf-8")


class WriteBatch:
    def __init__(self, ctx):
        self._ctx = ctx
        self._puts: Dict[str, Tuple[bytes, Optional[int]]] = {}
        self._deletes: Dict[str, None] = {}

    # ---- 入队 ----
    def put(self, key: str, obj: Any, expect: Optional[int] = None) -> None:
        self._deletes.pop(key, None)
        self._puts[key] = (encode(obj), expect)

    def put_bytes(self, key: str, value: bytes, expect: Optional[int] = None) -> None:
        self._deletes.pop(key, None)
        self._puts[key] = (value, expect)

    def delete(self, key: str) -> None:
        self._puts.pop(key, None)
        self._deletes[key] = None

    # ---- 统计 ----
    def size(self) -> int:
        return len(self._puts) + len(self._deletes)

    def operation_count(self) -> int:
        return sum(1 if expect is None else 2 for _value, expect in self._puts.values()) + len(
            self._deletes
        )

    # ---- 提交 ----
    def commit(self) -> int:
        sent = commit_batch(self._ctx, self)
        self._puts.clear()
        self._deletes.clear()
        return sent

    def __enter__(self):
        self._token = attach_batch(self)
        return self

    def __exit__(self, exc_type, exc, tb):
        try:
            if exc_type is None:
                self.commit()
        finally:
            _current_batch.reset(self._token)


def write_put(ctx, key: str, obj: Any, expect: Optional[int] = None) -> None:
    batch = current_batch()
    if batch is not None:
        batch.put(key, obj, expect)
        return
    one = WriteBatch(ctx)
    one.put(key, obj, expect)
    one.commit()


def write_put_bytes(ctx, key: str, value: bytes, expect: Optional[int] = None) -> None:
    batch = current_batch()
    if batch is not None:
        batch.put_bytes(key, value, expect)
        return
    one = WriteBatch(ctx)
    one.put_bytes(key, value, expect)
    one.commit()


def write_delete(ctx, key: str) -> None:
    batch = current_batch()
    if batch is not None:
        batch.delete(key)
        return
    one = WriteBatch(ctx)
    one.delete(key)
    one.commit()


def _is_unchanged(store, key: str, value: bytes, expect: Optional[int]) -> bool:
    raw = store.raw(key)
    if raw != value:
        return False
    if expect is None:
        return True
    if expect == CREATE:
        return False
    return store.mod_revision(key) == expect


def _refresh(transport, store, keys: Iterable[str]) -> None:
    for key in keys:
        value, meta = transport.get(key)
        store.apply_local(key, value, meta.mod_revision if meta is not None else 0)


def commit_batch(ctx, batch: WriteBatch) -> int:
    """把一批写按片提交，返回发出的操作数（便于测试与观测）。"""
    store = ctx.store
    transport = ctx.transport
    chunk_ops = max(10, int(ctx.chunk_ops))

    puts: List[Tuple[str, bytes, Optional[int]]] = []
    for key, (value, expect) in batch._puts.items():
        if _is_unchanged(store, key, value, expect):
            continue
        puts.append((key, value, expect))
    deletes: List[str] = [key for key in batch._deletes if True]

    # 分片：按操作数切
    chunks: List[Tuple[List[Tuple[str, bytes, Optional[int]]], List[str]]] = []
    cur_puts: List[Tuple[str, bytes, Optional[int]]] = []
    cur_deletes: List[str] = []
    cur_ops = 0
    for item in puts:
        cost = 1 if item[2] is None else 2
        if cur_ops + cost > chunk_ops and (cur_puts or cur_deletes):
            chunks.append((cur_puts, cur_deletes))
            cur_puts, cur_deletes, cur_ops = [], [], 0
        cur_puts.append(item)
        cur_ops += cost
    for key in deletes:
        if cur_ops + 1 > chunk_ops and (cur_puts or cur_deletes):
            chunks.append((cur_puts, cur_deletes))
            cur_puts, cur_deletes, cur_ops = [], [], 0
        cur_deletes.append(key)
        cur_ops += 1
    if cur_puts or cur_deletes:
        chunks.append((cur_puts, cur_deletes))

    sent_ops = 0
    for chunk_puts, chunk_deletes in chunks:
        _commit_chunk(ctx, store, transport, chunk_puts, chunk_deletes)
        sent_ops += sum(1 if e is None else 2 for _k, _v, e in chunk_puts) + len(chunk_deletes)
    return sent_ops


def _commit_chunk(ctx, store, transport, puts, deletes) -> None:
    attempts = 0
    while True:
        attempts += 1
        try:
            revision = 0
            if puts:
                revision = transport.txn_puts(puts)
                if not revision:
                    raise _Conflict()
            if deletes:
                ok = transport.txn_deletes(deletes)
                if not ok:
                    raise _Conflict()
        except (_Conflict, Conflict, EtcdError):
            if attempts >= ctx.cas_retries:
                raise Conflict(keys_label(puts, deletes), attempts)
            _refresh(transport, store, [k for k, _v, _e in puts] + list(deletes))
            puts = _rebuild_puts(store, puts)
            continue
        for key, value, _expect in puts:
            store.apply_local(key, value, revision)
        for key in deletes:
            store.apply_local(key, None)
        return


class _Conflict(Exception):
    pass


def _rebuild_puts(store, puts):
    rebuilt = []
    for key, value, expect in puts:
        if expect is None:
            rebuilt.append((key, value, None))
        elif expect == CREATE:
            rebuilt.append((key, value, CREATE))
        else:
            current = store.mod_revision(key)
            rebuilt.append((key, value, current if current is not None else expect))
    return rebuilt


def keys_label(puts, deletes) -> str:
    items = [k for k, _v, _e in puts] + list(deletes)
    return items[0] if items else "<empty>"
