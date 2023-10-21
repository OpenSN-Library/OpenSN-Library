import numpy as np
import cv2
from transmission_data import HeaderPacket, FramePacket, UDPPacket, build_udp_packet
from sender import UDPSender
from const_var import SEND_PORT, RECV_SERV_PORT, COMPRESS_SERV_PORT, FRAME_PER_PACKET
from loguru import logger
import time, uuid
from ip_translate import ip2str, str2ip
from arranger import get_compress_ip,refresh_ip_list

class VideoParser:
    def __init__(self, video_path: str,frame_array: list) -> None:
        self.video_path: str = video_path
        self.video_uuid: str = uuid.uuid4().hex
        self.current_frame: int = 0
        self.video = cv2.VideoCapture(self.video_path)
        self.total_frame_count: int = int(self.video.get(cv2.CAP_PROP_FRAME_COUNT))
        self.fps: int = int(self.video.get(cv2.CAP_PROP_FPS))
        self.fourcc: str = int(self.video.get(cv2.CAP_PROP_FOURCC)).to_bytes(4, byteorder="little").decode("utf-8")
        self.frame_size: tuple = (
            int(self.video.get(cv2.CAP_PROP_FRAME_WIDTH)), int(self.video.get(cv2.CAP_PROP_FRAME_HEIGHT)))
        self.frame_array = frame_array
    def get_frame_count(self) -> int:
        return self.total_frame_count

    def get_current_frame(self) -> int:
        return self.current_frame

    def get_next_frame(self) -> np.ndarray:
        exist, ret = self.video.read()
        ret = cv2.pyrDown(ret)
        ret = cv2.pyrDown(ret)
        ret = cv2.pyrDown(ret)
        ret = cv2.pyrDown(ret)
        self.current_frame += 1
        logger.info("Reading %d of %d"%(self.current_frame,self.get_frame_count()))
        # ret = cv2.imencode('.jpg', ret, [cv2.IMWRITE_JPEG_QUALITY, 80])
        self.frame_array.append(ret)
        return ret

    def reset_ptr(self) -> None:
        self.video.release()
        self.video = cv2.VideoCapture(self.video_path)



def video_producer_serv(video_path: str, target_ip: str,compress_ip:str, compress_rate: int, frame_array: list):
    refresh_ip_list()
    try:
        sender = UDPSender(SEND_PORT)
    except Exception as e:
        sender = UDPSender(SEND_PORT + 1)
    video = VideoParser(video_path,frame_array)
    while True:
        header_packet = HeaderPacket(
            total_frame_count=video.get_frame_count(),
            fps=video.fps,
            fourcc=video.fourcc,
            compress_rate=compress_rate,
            size=video.frame_size,
            video_uuid=video.video_uuid
        )
        head_udp_packet = build_udp_packet(header_packet)[0]
        logger.info("Sending Header...")
        sender.send_data(head_udp_packet.serialize(), [target_ip], RECV_SERV_PORT)
        frame_packets: list = [FramePacket(
            total_frame_count=video.get_frame_count(),
            start_frame_index=0,
            start_frame=video.get_next_frame(),
            compressed_rate=compress_rate,
            destination=str2ip(target_ip),
            video_uuid=video.video_uuid
        )]
        slice_counter = 1
        for i in range(1, video.get_frame_count()):
            if slice_counter > FRAME_PER_PACKET:
                udp_packets = build_udp_packet(frame_packets[-1])
                logger.info(
                    "Sending Frame %d ~ %d" % (
                        frame_packets[-1].start_frame_index,
                        frame_packets[-1].start_frame_index + frame_packets[-1].packet_frame_size - 1))
                valid_compress_ip = get_compress_ip(target_ip)
                for udp_packet in udp_packets:
                    logger.info("Sending Frame %d ~ %d of %d Slice %d" % (
                        frame_packets[-1].start_frame_index,
                        frame_packets[-1].start_frame_index + frame_packets[-1].packet_frame_size - 1,
                        video.get_frame_count(),
                        udp_packet.slice_index))
                    sender.send_data(udp_packet.serialize(), valid_compress_ip, COMPRESS_SERV_PORT)
                slice_counter = 1
                frame_packets.append(FramePacket(
                    total_frame_count=video.get_frame_count(),
                    start_frame_index=i,
                    start_frame=video.get_next_frame(),
                    compressed_rate=compress_rate,
                    destination=str2ip(target_ip),
                    video_uuid=video.video_uuid))
            else:
                frame_packets[-1].add_frame(video.get_next_frame())
            slice_counter += 1
        break
        # video.reset_ptr()
