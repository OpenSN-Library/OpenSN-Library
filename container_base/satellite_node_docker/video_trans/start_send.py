from video_produce import video_producer_serv
from const_var import VIDEO_PATH
from multiprocessing import Manager
import sys
from multiprocessing import Queue
from multiprocessing import Manager, Process
import websockets
import asyncio
import base64
import cv2
from loguru import logger

web_frame_array: list


async def echo(websocket,path):
    global web_frame_array
    index = len(web_frame_array) - 1
    while True:
        if len(web_frame_array) > 0 and index < len(web_frame_array):
            encoded, buffer = cv2.imencode('.jpg', web_frame_array[index], [cv2.IMWRITE_JPEG_QUALITY, 80])
            index += 1
            message = base64.b64encode(buffer)
            await websocket.send(message.decode())
            recv_ack = await websocket.recv()
            print(recv_ack)

if __name__ == "__main__":
    compress = sys.argv[1]
    recv = sys.argv[2]
    compress_level = 1
    video_path = VIDEO_PATH
    if len(sys.argv) > 3:
        video_path = sys.argv[3]
    if len(sys.argv) > 4:
        compress_level = int(sys.argv[4])
    web_frame_array = Manager().list()
    Process(
        target=video_producer_serv,
        args=(video_path, recv,compress, compress_level, web_frame_array)
    ).start()
    logger.info("Start Websocket...")
    while True:
        try:
            server = websockets.serve(echo, "0.0.0.0", 8765)
            asyncio.get_event_loop().run_until_complete(server)
            asyncio.get_event_loop().run_forever()
        except Exception as e:
            logger.warning("Websocket Dead... %s" % str(e))