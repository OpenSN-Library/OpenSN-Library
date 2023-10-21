import asyncio
import websockets
from multiprocessing import Process, Manager
from queue import Queue
import base64
import time
import cv2
import numpy as np


def display(packet_queue):
    i = 0
    while True:
        if not packet_queue.empty():
            packet = packet_queue.get()
            data = base64.b64decode(packet, ' /')
            npdata = np.frombuffer(data, dtype=np.uint8)
            frame = cv2.imdecode(npdata, 1)
            print(i)
            i += 1
            # cv2.imshow("WEBSOCKET_VIDEO", frame)
            # key = cv2.waitKey(1) & 0xFF
            time.sleep(0.016)


async def main(packet_queue):
    async with websockets.connect("ws://127.0.0.1:30010") as websocket:
        while True:
            packet = await websocket.recv()
            packet_queue.put(packet)



if __name__ == "__main__":
    queue = Manager().Queue()
    Process(target=display, args=(queue,)).start()
    asyncio.run(main(queue))
