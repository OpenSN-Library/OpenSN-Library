"""本地镜像：一次快照 + watch 增量（不可用时退化为轮询）。

循环里读实体都从这里拿，稳态 0 次 etcd 往返。
"""

import json
import threading
import time
from typing import Callable, Dict, List, Optional, Tuple

from . import keys
from .errors import NotFound, Unsupported

Entry = Tuple[bytes, int]  # (原始 JSON 字节, mod_revision)


class Change:
    def __init__(self, kind: str, key: str, value: Optional[dict]):
        self.kind = kind          # "put" / "delete"
        self.key = key
        self.value = value

    def __repr__(self) -> str:  # pragma: no cover - 调试用
        return f"Change({self.kind}, {self.key})"


class Store:
    def __init__(self, transport, prefixes=None, watch: bool = True, poll_interval: float = 2.0, extra_prefixes=None):
        self._transport = transport
        self._prefixes = list(prefixes or keys.SNAPSHOT_PREFIXES)
        for prefix in extra_prefixes or []:
            if prefix not in self._prefixes:
                self._prefixes.append(prefix)
        self._data: Dict[str, Entry] = {}
        self._lock = threading.RLock()
        self._watch = watch
        self._poll_interval = poll_interval
        self._stop = threading.Event()
        self._threads: List[threading.Thread] = []
        self._watchers: List = []
        self._revision = 0
        self._listeners: List[Callable[[Change], None]] = []
        self._changes: List[Change] = []

    # ---- 快照 ----
    def snapshot(self) -> None:
        data: Dict[str, Entry] = {}
        revision = 0
        for prefix in self._prefixes:
            items = self._transport.get_prefix(prefix)
            prefix_rev = 0
            for key, raw, meta in items:
                data[key] = (raw, meta.mod_revision)
                prefix_rev = max(prefix_rev, meta.mod_revision)
            revision = prefix_rev if revision == 0 else min(revision, prefix_rev)
        with self._lock:
            self._data = data
            self._revision = max(self._revision, revision)

    def start(self) -> None:
        if self._watch:
            if self._start_watch():
                return
        self._start_polling()

    def _start_watch(self) -> bool:
        try:
            for prefix in self._prefixes:
                watcher = self._transport.watch_prefix(prefix)
                self._watchers.append(watcher)
                t = threading.Thread(target=self._watch_loop, args=(prefix, watcher), daemon=True)
                t.start()
                self._threads.append(t)
            return True
        except Unsupported:
            self._stop_watchers()
            return False
        except Exception:
            self._stop_watchers()
            return False

    def _watch_loop(self, prefix: str, watcher) -> None:
        while not self._stop.is_set():
            try:
                for event in watcher:
                    if self._stop.is_set():
                        return
                    self._apply_event(event)
            except Exception:
                if self._stop.is_set():
                    return
                time.sleep(0.5)
                try:
                    watcher = self._transport.watch_prefix(prefix)
                except Exception:
                    return

    def _apply_event(self, event) -> None:
        key = getattr(event, "key", b"")
        if isinstance(key, bytes):
            key = key.decode("utf-8", "replace")
        value = getattr(event, "value", b"") or b""
        revision = int(getattr(event, "mod_revision", 0) or 0)
        if not key:
            return
        if value in (b"", None):
            with self._lock:
                existed = self._data.pop(key, None)
            if existed is not None:
                self._emit(Change("delete", key, None))
        else:
            with self._lock:
                old = self._data.get(key)
                self._data[key] = (value, revision)
                self._revision = max(self._revision, revision)
            if old is None or old[0] != value:
                self._emit(Change("put", key, _loads(value)))

    def _start_polling(self) -> None:
        t = threading.Thread(target=self._poll_loop, daemon=True)
        t.start()
        self._threads.append(t)

    def _poll_loop(self) -> None:
        while not self._stop.is_set():
            time.sleep(self._poll_interval)
            try:
                self.refresh()
            except Exception:
                continue

    def refresh(self) -> None:
        """重新拉一遍所有前缀（watch 断线或轮询模式用）。"""
        for prefix in self._prefixes:
            items = self._transport.get_prefix(prefix)
            current = {}
            for key, raw, meta in items:
                current[key] = (raw, meta.mod_revision)
            with self._lock:
                for key in list(self._data):
                    if key.startswith(prefix) and key not in current:
                        self._data.pop(key, None)
                        self._emit(Change("delete", key, None))
                for key, entry in current.items():
                    old = self._data.get(key)
                    self._data[key] = entry
                    if old is None or old[0] != entry[0]:
                        self._emit(Change("put", key, _loads(entry[0])))
                    self._revision = max(self._revision, entry[1])

    # ---- 读 ----
    def raw(self, key: str) -> Optional[bytes]:
        with self._lock:
            entry = self._data.get(key)
            return None if entry is None else entry[0]

    def obj(self, key: str) -> Optional[dict]:
        raw = self.raw(key)
        return None if raw is None else _loads(raw)

    def require(self, key: str, entity: str) -> dict:
        data = self.obj(key)
        if data is None:
            raise NotFound(entity, key)
        return data

    def mod_revision(self, key: str) -> Optional[int]:
        with self._lock:
            entry = self._data.get(key)
            return None if entry is None else entry[1]

    def keys(self, prefix: str) -> List[str]:
        with self._lock:
            return [k for k in self._data if k.startswith(prefix)]

    def items(self, prefix: str) -> List[Tuple[str, dict]]:
        out = []
        with self._lock:
            snapshot = [(k, v) for k, v in self._data.items() if k.startswith(prefix)]
        for key, (raw, _rev) in snapshot:
            out.append((key, _loads(raw)))
        return out

    # ---- 写后乐观更新 ----
    def apply_local(self, key: str, value: Optional[bytes], revision: int = 0) -> None:
        with self._lock:
            if value is None:
                self._data.pop(key, None)
            else:
                self._data[key] = (value, int(revision or self._revision or 1))

    # ---- 变更订阅 ----
    def on_change(self, callback: Callable[[Change], None]) -> None:
        self._listeners.append(callback)

    def _emit(self, change: Change) -> None:
        self._changes.append(change)
        if len(self._changes) > 4096:
            self._changes = self._changes[-2048:]
        for callback in list(self._listeners):
            try:
                callback(change)
            except Exception:
                pass

    def drain_changes(self) -> List[Change]:
        with self._lock:
            changes = self._changes
            self._changes = []
        return changes

    # ---- 统计与关闭 ----
    def cache_size(self) -> int:
        with self._lock:
            return sum(len(raw) + len(key) + 16 for key, (raw, _rev) in self._data.items())

    def close(self) -> None:
        self._stop.set()
        self._stop_watchers()
        for t in self._threads:
            t.join(timeout=1.0)
        self._threads = []

    def _stop_watchers(self) -> None:
        for watcher in self._watchers:
            try:
                watcher.close()
            except Exception:
                pass
        self._watchers = []


def _loads(raw: bytes) -> dict:
    try:
        return json.loads(raw.decode("utf-8"))
    except Exception:
        return {}
