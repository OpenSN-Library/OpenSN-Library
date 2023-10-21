import time

import numpy as np
from transmission_data import FramePacket, HeaderPacket, build_udp_packet
from sender import UDPSender
from const_var import COMPRESS_SEND_PORT, RECV_SERV_PORT
from ip_translate import ip2str, str2ip
import cv2



def compress_image(frame: np.ndarray, n: int = 1) -> np.ndarray:
    res = frame
    for i in range(n):
        time.sleep(0.5)
        res = cv2.pyrDown(res)
    return res


def video_compresser(frame: FramePacket = None, header: HeaderPacket = None):
    udp_sender = UDPSender(COMPRESS_SEND_PORT)

    for i in range(len(frame.frames)):
        frame.frames[i] = compress_image(frame.frames[i],frame.compressed_rate)
    frame.frame_size = (
        frame.frame_size[0] // (1 << frame.compressed_rate),
        frame.frame_size[1] // (1 << frame.compressed_rate)
    )
    udp_packets = build_udp_packet(frame)
    for packet in udp_packets:
        udp_sender.send_data(packet.serialize(),[ip2str(frame.destination)] , RECV_SERV_PORT)


