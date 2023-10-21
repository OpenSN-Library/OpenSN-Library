import asyncio
import websockets
import base64
import time
import cv2
import numpy as np

fps, st, frames_to_count, cnt = (0, 0, 20, 0)


async def main():

    global fps, st, frames_to_count, cnt
    async with websockets.connect("ws://localhost:8765") as websocket:
        while True:
            packet = await websocket.recv()
            data = base64.b64decode(packet, ' /')
            # data_str = data.decode()
            npdata = np.frombuffer(data, dtype=np.uint8)
            # print(npdata)
            frame = cv2.imdecode(npdata, 1)
            frame = cv2.putText(frame, 'FPS: ' + str(fps), (10, 40), cv2.FONT_HERSHEY_SIMPLEX, 0.7, (0, 0, 255), 2)
            print(frame)
            cv2.imshow("RECEIVING_VIDEO", frame)
            key = cv2.waitKey(1) & 0xFF
            # print("b")
            if cnt == frames_to_count:
                try:
                    fps = round(frames_to_count / (time.time() - st))
                    st = time.time()
                    cnt = 0
                except:
                    pass
            cnt += 1


asyncio.run(main())
