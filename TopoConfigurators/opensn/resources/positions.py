"""位置（弧度 / 米）。daemon 不消费，给前端展示和其他配置器读。"""

from typing import Dict, List

from .. import keys
from ..errors import NotFound
from ..schema.entities import Position
from ..write import write_delete, write_put


class Positions:
    def __init__(self, ctx):
        self._ctx = ctx

    def get(self, instance_id) -> Position:
        instance_id = getattr(instance_id, "instance_id", instance_id)
        key = keys.position(str(instance_id))
        raw = self._ctx.store.obj(key)
        if raw is None:
            raise NotFound("position", key)
        return Position.from_json(raw)

    def map(self) -> Dict[str, Position]:
        out = {}
        for key, raw in self._ctx.store.items(keys.POSITION + "/"):
            out[keys.split_last(key)] = Position.from_json(raw)
        return out

    def set(self, instance, position: Position) -> None:
        instance_id = getattr(instance, "instance_id", instance)
        write_put(self._ctx, keys.position(str(instance_id)), position)

    def set_all(self, mapping: Dict[str, Position]) -> None:
        for instance, position in mapping.items():
            self.set(instance, position)

    def delete(self, instance) -> None:
        instance_id = getattr(instance, "instance_id", instance)
        write_delete(self._ctx, keys.position(str(instance_id)))

    def delete_all(self) -> int:
        keys_ = list(self._ctx.store.keys(keys.POSITION + "/"))
        for key in keys_:
            write_delete(self._ctx, key)
        return len(keys_)
