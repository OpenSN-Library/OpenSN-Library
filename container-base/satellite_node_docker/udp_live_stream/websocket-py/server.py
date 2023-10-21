import asyncio
import websockets
import base64
import time
import cv2
import imutils

WIDTH = 400
vid = cv2.VideoCapture("out.mp4")  # replace 'rocket.mp4' with 0 for webcam
fps, st, frames_to_count, cnt = (0, 0, 20, 0)


async def echo(websocket, path):
    global fps, st, frames_to_count, cnt
    while (vid.isOpened()):
        _, frame = vid.read()
        frame = imutils.resize(frame, width=WIDTH)
        encoded, buffer = cv2.imencode('.jpg', frame, [cv2.IMWRITE_JPEG_QUALITY, 80])
        message = base64.b64encode(buffer)
        # print(message.decode())
        await websocket.send(message.decode())
        frame = cv2.putText(frame, 'FPS: ' + str(fps), (10, 40), cv2.FONT_HERSHEY_SIMPLEX, 0.7, (0, 0, 255), 2)
        cv2.imshow('TRANSMITTING VIDEO', frame)
        key = cv2.waitKey(1) & 0xFF
        if cnt == frames_to_count:
            try:
                fps = round(frames_to_count / (time.time() - st))
                st = time.time()
                cnt = 0
            except:
                pass
        cnt += 1


start_server = websockets.serve(echo, "localhost", 8765)

asyncio.get_event_loop().run_until_complete(start_server)
asyncio.get_event_loop().run_forever()
