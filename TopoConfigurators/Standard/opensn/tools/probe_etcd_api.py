#!/usr/bin/env python3
"""在真实 etcd 上探测 python-etcd3 的能力。

首次部署时先跑一次；确认事务/监听接口与 SDK 预期一致::

    OPENSN_ETCD_ENDPOINTS=10.0.0.1:2379 python3 opensn/tools/probe_etcd_api.py
"""

import sys
import time

from opensn.transport import CREATE, EtcdTransport


def main() -> int:
    transport = EtcdTransport()
    key = "/opensn_probe/keys"
    try:
        value, meta = transport.get("/nodes")
        print("get:", "ok" if value is not None or meta is None else "ok(empty)")
    except Exception as exc:
        print("get FAILED:", exc)
        return 1

    try:
        transport.put(key, b"1")
        print("put: ok")
    except Exception as exc:
        print("put FAILED:", exc)
        return 1

    try:
        revision = transport.txn_puts([(key, b"2", CREATE)])
        print("txn_puts(create) ->", revision)
    except Exception as exc:
        print("txn_puts FAILED:", exc)
        print("提示：若客户端确实不支持事务，可设 OPENSN_ETCD_COMPAT=1 退化运行（非原子）")
        return 1

    try:
        _, meta = transport.get(key)
        revision = transport.txn_puts([(key, b"3", meta.mod_revision if meta else None)])
        print("txn_puts(mod) ->", revision)
    except Exception as exc:
        print("txn_puts(mod) FAILED:", exc)
        return 1

    try:
        watcher = transport.watch_prefix(key)
        print("watch_prefix: ok")
        transport.put(key, b"4")
        deadline = time.time() + 2
        got = False
        for _event in watcher:
            got = True
            break
        print("watch event:", "ok" if got else "timeout")
        watcher.close()
    except Exception as exc:
        print("watch_prefix FAILED:", exc, "（SDK 可设置 watch=False 退化为轮询）")

    transport.delete(key)
    transport.delete_prefix("/opensn_probe")
    transport.close()
    print("probe done")
    return 0


if __name__ == "__main__":
    sys.exit(main())
