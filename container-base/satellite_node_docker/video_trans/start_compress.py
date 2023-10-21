from receiver import UDPRecevier, video_constructor, CONSTRUCT_TYPE, COMPRESS_TYPE
import threading
from video_encode import video_constructor, VideoEncoder, video_store,display_frame
from loguru import logger
from const_var import COMPRESS_SERV_PORT
import time
import cv2


if __name__ == "__main__":
    recv = UDPRecevier(COMPRESS_SERV_PORT, COMPRESS_TYPE, None)
    while True:
        if not recv.check_alive():
            logger.info("Receiver Process Dead, Restarting...")
            recv.close()
            recv = UDPRecevier(COMPRESS_SERV_PORT, COMPRESS_TYPE, None)
        time.sleep(5)