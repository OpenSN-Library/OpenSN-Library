from receiver import UDPRecevier, CONSTRUCT_TYPE
from loguru import logger
from const_var import RECV_SERV_PORT
import cv2
from multiprocessing import Queue
from multiprocessing import Manager
import websockets
import asyncio
import base64

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
    array = Manager().list()
    web_frame_array = array
    recv = UDPRecevier(RECV_SERV_PORT, CONSTRUCT_TYPE, array)
    logger.info("Start Websocket...")
    while True:
        try:
            server = websockets.serve(echo, "0.0.0.0", 8765)
            asyncio.get_event_loop().run_until_complete(server)
            asyncio.get_event_loop().run_forever()
        except Exception as e:
            logger.warning("Websocket Dead... %s" % str(e))
