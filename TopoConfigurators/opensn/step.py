"""步骤循环：把"提交上一轮写入 + 等到下一个时间点 + 给本轮快照"收在一起。"""

import time
from typing import Optional

from . import keys
from .write import WriteBatch, attach_batch, reset_batch


class StepPositions:
    def __init__(self, batch: WriteBatch):
        self._batch = batch

    def set(self, instance, position) -> None:
        instance_id = getattr(instance, "instance_id", instance)
        self._batch.put(keys.position(str(instance_id)), position)

    def set_all(self, mapping) -> None:
        for instance, position in mapping.items():
            self.set(instance, position)


class Step:
    def __init__(self, ctx, batch: WriteBatch, interval: float):
        self._ctx = ctx
        self._batch = batch
        self.interval = interval
        self.now = time.time()
        self.positions = StepPositions(batch)

    @property
    def instances(self):
        return self._ctx.instances

    @property
    def links(self):
        return self._ctx.links

    def flush(self) -> int:
        return self._batch.commit()


class Steps:
    def __init__(self, ctx, interval: float = 5.0, delay_quantum_s: Optional[float] = None):
        self._ctx = ctx
        self._interval = float(interval)
        self._delay_quantum_s = delay_quantum_s
        self._batch: Optional[WriteBatch] = None
        self._token = None
        self._next_at = time.monotonic()
        self._closed = False

    def __iter__(self):
        return self

    def __next__(self) -> Step:
        if self._closed:
            raise StopIteration
        if self._batch is not None:
            self._batch.commit()
            reset_batch(self._token)
            self._next_at += self._interval
            delay = self._next_at - time.monotonic()
            if delay > 0:
                time.sleep(delay)
            else:
                self._next_at = time.monotonic()
        self._ctx.delay_quantum_s = self._delay_quantum_s
        self._batch = WriteBatch(self._ctx)
        self._token = attach_batch(self._batch)
        return Step(self._ctx, self._batch, self._interval)

    def close(self) -> None:
        if self._closed:
            return
        self._closed = True
        if self._batch is not None:
            self._batch.commit()
            self._batch = None
        if self._token is not None:
            reset_batch(self._token)
            self._token = None

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc, tb):
        self.close()
