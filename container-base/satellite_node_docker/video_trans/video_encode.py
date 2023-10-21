from collections import OrderedDict
from transmission_data import FramePacket, HeaderPacket
import numpy as np
from loguru import logger
import asyncio
import websockets
import base64
import cv2
import time
import datetime

class VideoEncoder():

    def __init__(self):
        self.frame_dict = OrderedDict()
        self.frame_count = -1
        self.received_frame_count = 0
        self.fps = -1
        self.fourcc = "NONE"
        self.size = (-1, -1)
        self.start_recv_time = datetime.datetime.now()
        self.recv_end_time = 0

    def set_header(self, header: HeaderPacket):
        self.frame_count = header.total_frame_count
        self.fps = header.fps
        self.fourcc = header.fourcc
        self.size = header.size

    def add_frame(self, frame: np.ndarray, frame_index: int):
        self.frame_dict[frame_index] = frame
        self.received_frame_count += 1

    def receive_finished(self):
        logger.info("Received %d of %d" % (self.received_frame_count, self.frame_count))
        return 0 < self.frame_count <= self.received_frame_count + 1

    def encode_to(self, path: str):
        encodeer = cv2.VideoWriter(path, cv2.VideoWriter_fourcc(*self.fourcc), self.fps, self.size)
        for frame in self.frame_dict.values():
            encodeer.write(frame)
        encodeer.release()


video_store: dict = {}
web_frame_array: list


def display_frame(dis_frame_array):
    logger.info("Start Displaying...")
    index = 0
    while True:
        if index < len(dis_frame_array):
            cv2.imshow('TRANSMITTING VIDEO', dis_frame_array[index])
            cv2.waitKey(1) & 0xFF
            time.sleep(0.016)
            index += 1


def video_constructor(frame: FramePacket = None, header: HeaderPacket = None, frame_array=[]):
    global video_store
    if frame is not None:
        video_id = frame.video_uuid
        if frame.video_uuid not in video_store.keys():
            logger.warning("New video: %s" % frame.video_uuid)
            video_store[frame.video_uuid] = VideoEncoder()
        for i in range(frame.start_frame_index, frame.start_frame_index + len(frame.frames)):
            video_store[frame.video_uuid].add_frame(frame.frames[i - frame.start_frame_index], i)
            frame_array.append(frame.frames[i - frame.start_frame_index])
            if frame.frames[i - frame.start_frame_index].shape[0] == 1:
                print("EN")
            print("Add Frame %d : %s" % (i, frame.frames[i - frame.start_frame_index].shape))
    elif header is not None:
        video_id = header.video_uuid
        if header.video_uuid not in video_store.keys():
            video_store[header.video_uuid] = VideoEncoder()
        video_store[header.video_uuid].set_header(header)
    else:
        logger.warning("None Built")
        return None
    if video_store[video_id].receive_finished():
        # video_store[video_id].encode_to(video_id + ".mp4")
        video_store[video_id].recv_end_time = datetime.datetime.now()
        logger.info("Receive Time is %d second"%(video_store[video_id].recv_end_time-video_store[video_id].start_recv_time).seconds)
        video_store.pop(video_id)
