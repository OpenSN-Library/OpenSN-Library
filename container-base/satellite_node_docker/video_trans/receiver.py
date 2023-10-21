import socket
from queue import Queue

from transmission_data import UDPPacket, HEAD_PACKET_TYPE, UDP_PACKET_SIZE_THRESHOLD, HeaderPacket, FramePacket
from video_encode import video_constructor, display_frame
from video_compresser import video_compresser
from loguru import logger
from collections import OrderedDict
from multiprocessing import Manager, Process

CONSTRUCT_TYPE = 1
COMPRESS_TYPE = 2


def build_frame(work_type, pipe, array):
    frame_store: dict = {}
    frame_slice_count: dict = {}
    while True:
        if not pipe.empty():
            data_recv = pipe.get()
            recv_packet = UDPPacket()
            recv_packet.deserialize(data_recv)
            if recv_packet.packet_id not in frame_store.keys():
                logger.info("Receiving Packet: %s" % recv_packet.packet_id)
                frame_store[recv_packet.packet_id] = OrderedDict()
                frame_slice_count[recv_packet.packet_id] = recv_packet.slice_count
            frame_store[recv_packet.packet_id][recv_packet.slice_index] = recv_packet
            if len(frame_store[recv_packet.packet_id]) >= frame_slice_count[recv_packet.packet_id]:
                packet_type = frame_store[recv_packet.packet_id][0].data_bytes[0]
                byte_seq_full = b''
                for key, packet in frame_store[recv_packet.packet_id].items():
                    byte_seq_full += packet.data_bytes
                frame_store.pop(recv_packet.packet_id)
                if packet_type == HEAD_PACKET_TYPE:
                    header_packet = HeaderPacket()
                    frame_packet = None
                    header_packet.deserialize(byte_seq_full)
                else:
                    frame_packet = FramePacket()
                    header_packet = None
                    frame_packet.deserialize(byte_seq_full)
                if work_type == COMPRESS_TYPE:
                    video_compresser(frame_packet, header_packet)
                elif work_type == CONSTRUCT_TYPE:
                    video_constructor(frame_packet, header_packet, array)


def udp_recv_handler(socket_fd: socket.socket, work_type: int, array: list):
    logger.info("Start UDP Receiving...")
    manager = Manager()
    queue = manager.Queue()
    Process(target=build_frame, args=(work_type, queue, array)).start()
    if work_type == CONSTRUCT_TYPE:
        # Process(target=display_frame, args=(array,)).start()    
        pass
    while True:
        data_recv, addr = socket_fd.recvfrom(UDP_PACKET_SIZE_THRESHOLD + 1000)
        queue.put(data_recv)
        socket_fd.sendto(b"ACK", addr)


class UDPRecevier:

    def __create_receiver_socket(self, port: int) -> socket.socket:
        s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        s.bind(('', port))
        return s

    def __init__(self, port: int, work_type: int,queue: Queue):
        self.port: int = port
        self.socket: socket.socket = self.__create_receiver_socket(port)
        self.work_type = work_type
        self.process = Process(
            target=udp_recv_handler,
            args=(self.socket, self.work_type, queue)
        )
        self.process.start()

    def check_alive(self) -> bool:
        return self.process.is_alive()

    def close(self):
        self.process.terminate()
        self.socket.close()
