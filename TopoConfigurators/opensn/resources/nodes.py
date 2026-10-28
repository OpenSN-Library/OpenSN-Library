"""节点（只读）。"""

from typing import Dict, List

from .. import keys
from ..errors import NotFound
from ..schema.entities import Node


class Nodes:
    def __init__(self, ctx):
        self._ctx = ctx

    def list(self) -> List[Node]:
        out = []
        for key, raw in self._ctx.store.items(keys.NODES + "/"):
            try:
                out.append(Node.from_json(raw))
            except Exception:
                continue
        out.sort(key=lambda n: n.node_index)
        return out

    def get(self, node_index: int) -> Node:
        key = keys.node(node_index)
        raw = self._ctx.store.obj(key)
        if raw is None:
            raise NotFound("node", key)
        return Node.from_json(raw)

    def map(self) -> Dict[int, Node]:
        return {node.node_index: node for node in self.list()}
