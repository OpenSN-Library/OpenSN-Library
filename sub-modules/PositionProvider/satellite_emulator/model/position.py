import json
from satellite_emulator.utils.tools import object2dict

class Position:

    def __init__(self):
        self.latitude: float = 0
        self.longitude: float = 0
        self.altitude: float = 0

def position_from_json(seq: str) -> Position:
    position = Position()
    position.__dict__ = json.loads(seq)
    return position

def position_to_json(position: Position) -> str:
    return json.dumps(position.__dict__)