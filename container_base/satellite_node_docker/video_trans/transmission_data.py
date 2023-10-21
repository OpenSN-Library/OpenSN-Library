import uuid
import numpy as np
import math
from loguru import logger

HEAD_PACKET_TYPE = 1
FRAME_PACKET_TYPE = 2
UDP_PACKET_TYPE = 3
UDP_PACKET_SIZE_THRESHOLD = 8192


class SerializiblePacket:

    def serialize(self) -> bytes:
        pass

    def deserialize(self, data: bytes):
        pass


class HeaderPacket(SerializiblePacket):

    def __init__(
            self,
            total_frame_count: int = -1,
            fourcc: str = "NONE",
            size: tuple = (-1, -1),
            compress_rate: int = 0,
            fps: int = -1,
            video_uuid: str = "NONE"
    ):
        self.packet_type = HEAD_PACKET_TYPE
        self.video_uuid = video_uuid
        self.total_frame_count = total_frame_count
        self.fourcc = fourcc
        self.size = size
        self.compress_rate = compress_rate
        self.fps = fps

    def serialize(self) -> bytes:
        byte_seq = self.packet_type.to_bytes(1, byteorder='big')  # 1 byte
        byte_seq += self.video_uuid.encode()  # 32 bytes
        byte_seq += self.total_frame_count.to_bytes(4, byteorder='big')  # 4 bytes
        byte_seq += self.fourcc.encode("utf-8")  # 4 bytes
        byte_seq += self.size[0].to_bytes(4, byteorder='big')  # 4 bytes
        byte_seq += self.size[1].to_bytes(4, byteorder='big')  # 4 bytes
        byte_seq += self.compress_rate.to_bytes(4, byteorder='big')  # 4 bytes
        byte_seq += self.fps.to_bytes(4, byteorder='big')  # 4 bytes
        return byte_seq

    def deserialize(self, byte_seq: bytes):
        self.packet_type = int.from_bytes(byte_seq[0:1], byteorder='big')
        self.video_uuid = byte_seq[1:33].decode("utf-8")
        self.total_frame_count = int.from_bytes(byte_seq[33:37], byteorder='big')
        self.fourcc = byte_seq[37:41].decode("utf-8")
        self.size = (int.from_bytes(byte_seq[41:45], byteorder='big'), int.from_bytes(byte_seq[45:49], byteorder='big'))
        self.compress_rate = int.from_bytes(byte_seq[49:53], byteorder='big')
        self.fps = int.from_bytes(byte_seq[53:57], byteorder='big')

    def get_packet_byte_size(self) -> int:
        return 57


class FramePacket(SerializiblePacket):

    def __init__(
            self,
            total_frame_count: int = -1,
            start_frame_index: int = -1,
            destination: int = 0,
            compressed_rate: int = 1,
            start_frame: np.ndarray = None,
            video_uuid: str = "NONE"
    ):
        if start_frame is None:
            start_frame = []
        self.packet_type = FRAME_PACKET_TYPE
        self.video_uuid = video_uuid
        self.total_frame_count = total_frame_count
        self.start_frame_index = start_frame_index
        self.packet_frame_size = 1
        self.destination = destination
        self.compressed = False
        self.compressed_rate = compressed_rate
        if len(start_frame) <= 0:
            self.frames = []
            self.frame_size = (0, 0)
        else:
            self.frames: list[np.ndarray] = [start_frame]
            self.frame_size = self.frames[0].shape

    def add_frame(self, frame: np.ndarray):
        self.frames.append(frame)
        self.packet_frame_size += 1

    def serialize(self):
        byte_seq = self.packet_type.to_bytes(1, byteorder='big')
        byte_seq += self.video_uuid.encode("utf-8")
        byte_seq += self.total_frame_count.to_bytes(4, byteorder='big')
        byte_seq += self.start_frame_index.to_bytes(4, byteorder='big')
        byte_seq += self.packet_frame_size.to_bytes(4, byteorder='big')
        byte_seq += self.destination.to_bytes(4, byteorder='big')
        byte_seq += self.compressed.to_bytes(1, byteorder='big')
        byte_seq += self.compressed_rate.to_bytes(4, byteorder='big')
        byte_seq += self.frame_size[0].to_bytes(4, byteorder='big')
        byte_seq += self.frame_size[1].to_bytes(4, byteorder='big')
        for frame in self.frames:
            byte_seq += frame.tobytes()
        return byte_seq

    def deserialize(self, byte_seq: bytes):
        self.packet_type = int.from_bytes(byte_seq[0:1], byteorder='big')
        self.video_uuid = byte_seq[1:33].decode("utf-8")
        self.total_frame_count = int.from_bytes(byte_seq[33:37], byteorder='big')
        self.start_frame_index = int.from_bytes(byte_seq[37:41], byteorder='big')
        self.packet_frame_size = int.from_bytes(byte_seq[41:45], byteorder='big')
        self.destination = int.from_bytes(byte_seq[45:49], byteorder='big')
        self.compressed = bool.from_bytes(byte_seq[49:50], byteorder='big')
        self.compressed_rate = int.from_bytes(byte_seq[50:54], byteorder='big')
        self.frame_size = (
        int.from_bytes(byte_seq[54:58], byteorder='big'), int.from_bytes(byte_seq[58:62], byteorder='big'))
        for i in range(self.packet_frame_size):
            self.frames.append(
                np.frombuffer(
                    byte_seq[62 + i * self.frame_size[0] * self.frame_size[1] * 3:
                             62 + (i + 1) * self.frame_size[0] * self.frame_size[1] * 3],
                    dtype=np.uint8).reshape((self.frame_size[0], self.frame_size[1], 3)))

    def get_packet_byte_size(self) -> int:
        return 62 + self.packet_frame_size * self.frame_size[0] * self.frame_size[1]


class UDPPacket:

    def __init__(self, total_slice: int = 0, slice_index: int = 0, data_bytes: bytes = b'', packet_uuid: str = None):
        self.packet_type = UDP_PACKET_TYPE
        self.packet_id = packet_uuid
        self.slice_count = total_slice
        self.slice_index = slice_index
        self.data_bytes = data_bytes

    def serialize(self) -> bytes:
        data_bytes = self.packet_type.to_bytes(1, byteorder='big')
        data_bytes += self.packet_id.encode("utf-8")
        data_bytes += self.slice_count.to_bytes(4, byteorder='big')
        data_bytes += self.slice_index.to_bytes(4, byteorder='big')
        data_bytes += self.data_bytes
        return data_bytes

    def deserialize(self, data_bytes):
        self.packet_type = int.from_bytes(data_bytes[0:1], byteorder='big')
        self.packet_id = data_bytes[1:33].decode("utf-8")
        self.slice_count = int.from_bytes(data_bytes[33:37], byteorder='big')
        self.slice_index = int.from_bytes(data_bytes[37:41], byteorder='big')
        self.data_bytes = data_bytes[41:]

    def get_packet_byte_size(self) -> int:
        return 41 + len(self.data_bytes)


def build_udp_packet(raw_packet: SerializiblePacket) -> list:
    if isinstance(raw_packet, FramePacket):
        logger.info("Building UDP Packet of Frame Set %d..." % raw_packet.start_frame_index)
    packet_uuid = uuid.uuid4().hex
    packet_byte_seq = raw_packet.serialize()
    packet_byte_size = len(packet_byte_seq)
    slice_count = math.ceil(packet_byte_size / UDP_PACKET_SIZE_THRESHOLD)
    udp_packets: list = []
    for i in range(slice_count):
        right = min((i + 1) * UDP_PACKET_SIZE_THRESHOLD, packet_byte_size)
        udp_packets.append(UDPPacket(slice_count, i, packet_byte_seq[i * UDP_PACKET_SIZE_THRESHOLD:right], packet_uuid))
    logger.info("Build UDP Packet %d %d" % (slice_count, len(udp_packets)))
    return udp_packets
