"""内存 etcd 桩：只实现 SDK 用到的接口，供单元测试使用。"""

import queue
import threading
import time


class Meta:
    def __init__(self, key, mod_revision, create_revision=0, version=1):
        self.key = key.encode("utf-8") if isinstance(key, str) else key
        self.mod_revision = mod_revision
        self.create_revision = create_revision
        self.version = version


class Event:
    def __init__(self, key: str, value: bytes, mod_revision: int):
        self.key = key.encode("utf-8")
        self.value = value
        self.mod_revision = mod_revision


class _Compare:
    def __init__(self, kind: str, key: str):
        self.kind = kind
        self.key = key

    def __eq__(self, other):  # txns.create(key) == 0
        return _CompareValue(self.kind, self.key, other)


class _CompareValue:
    def __init__(self, kind, key, value):
        self.kind = kind
        self.key = key
        self.value = value


class _Op:
    def __init__(self, kind, key, value=None):
        self.kind = kind
        self.key = key
        self.value = value


class _Transactions:
    def create(self, key):
        return _Compare("create", key)

    def mod(self, key):
        return _Compare("mod", key)

    def put(self, key, value):
        return _Op("put", key, value if isinstance(value, bytes) else str(value).encode("utf-8"))

    def delete(self, key):
        return _Op("delete", key)


class _Header:
    def __init__(self, revision):
        self.revision = revision


class _Response:
    def __init__(self, succeeded, revision):
        self.succeeded = succeeded
        self.header = _Header(revision)


class FakeEtcd:
    def __init__(self):
        self._data = {}  # key -> [bytes, mod_rev, create_rev, version]
        self._revision = 0
        self.lock = threading.RLock()
        self._watchers = []
        self.transactions = _Transactions()
        self.txn_count = 0
        self.put_count = 0

    # ---- 基础 ----
    def _bump(self):
        self._revision += 1
        return self._revision

    def get(self, key):
        with self.lock:
            entry = self._data.get(key)
            if entry is None:
                return None, None
            return entry[0], Meta(key, entry[1], entry[2], entry[3])

    def get_prefix(self, prefix):
        with self.lock:
            keys = sorted(k for k in self._data if k.startswith(prefix))
            return [(self._data[k][0], Meta(k, self._data[k][1], self._data[k][2], self._data[k][3])) for k in keys]

    def put(self, key, value):
        self.txn_count += 1
        self.put_count += 1
        with self.lock:
            rev = self._bump()
            entry = self._data.get(key)
            if entry is None:
                self._data[key] = [value, rev, rev, 1]
            else:
                entry[0] = value
                entry[1] = rev
                entry[3] += 1
            self._notify(key, value, rev)
        return Meta(key, rev)

    def delete(self, key):
        self.txn_count += 1
        with self.lock:
            entry = self._data.pop(key, None)
            if entry is not None:
                rev = self._bump()
                self._notify(key, b"", rev)

    def delete_prefix(self, prefix):
        with self.lock:
            for key in [k for k in self._data if k.startswith(prefix)]:
                self.delete(key)

    def close(self):
        pass

    # ---- 事务 ----
    def transaction(self, compare=None, success=None, failure=None):
        self.txn_count += 1
        with self.lock:
            ok = all(self._check(item) for item in (compare or []))
            ops = success if ok else (failure or [])
            for op in ops or []:
                if op.kind == "put":
                    self._apply_put(op.key, op.value)
                elif op.kind == "delete":
                    self._apply_delete(op.key)
            return _Response(ok, self._revision)

    def _check(self, item):
        if not isinstance(item, _CompareValue):
            return False
        entry = self._data.get(item.key)
        if item.kind == "create":
            return entry is None
        if item.kind == "mod":
            return entry is not None and entry[1] == item.value
        return False

    def _apply_put(self, key, value):
        rev = self._bump()
        entry = self._data.get(key)
        if entry is None:
            self._data[key] = [value, rev, rev, 1]
        else:
            entry[0] = value
            entry[1] = rev
            entry[3] += 1
        self._notify(key, value, rev)

    def _apply_delete(self, key):
        entry = self._data.pop(key, None)
        if entry is not None:
            self._notify(key, b"", self._bump())

    # ---- 监听 ----
    def _notify(self, key, value, revision):
        for prefix, q in list(self._watchers):
            if key.startswith(prefix):
                q.put(Event(key, value, revision))

    def watch_prefix(self, prefix):
        q = queue.Queue()
        with self.lock:
            self._watchers.append((prefix, q))
        cancelled = threading.Event()

        def generator():
            while not cancelled.is_set():
                try:
                    event = q.get(timeout=0.1)
                except queue.Empty:
                    continue
                yield event

        def cancel():
            cancelled.set()
            with self.lock:
                self._watchers = [(p, qq) for (p, qq) in self._watchers if qq is not q]

        return generator(), cancel

    # ---- 测试辅助 ----
    def seed(self, key, value):
        """直接写（不算事务），用于搭初始状态。"""
        with self.lock:
            rev = self._bump()
            entry = self._data.get(key)
            if entry is None:
                self._data[key] = [value if isinstance(value, bytes) else str(value).encode(), rev, rev, 1]
            else:
                entry[0] = value if isinstance(value, bytes) else str(value).encode()
                entry[1] = rev
                entry[3] += 1
            self._notify(key, self._data[key][0], rev)

    def dump(self, prefix=""):
        with self.lock:
            return {k: v[0] for k, v in self._data.items() if k.startswith(prefix)}
