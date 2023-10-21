# This is client code to receive video frames over UDP
import base64
import time
import cv2
import numpy as np
import socket

BUFF_SIZE = 65536
client_socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
client_socket.setsockopt(socket.SOL_SOCKET, socket.SO_RCVBUF, BUFF_SIZE)
client_socket.setsockopt(socket.SOL_SOCKET, socket.SO_DEBUG, 1)
client_socket.setsockopt(socket.IPPROTO_IP, socket.IP_OPTIONS, bytearray([0x93, 0x6, 0x1, 0x2, 0x3, 0x4]))
host_name = socket.gethostname()
host_ip = '172.18.0.1'  # socket.gethostbyname(host_name)
print(host_ip)
port = 9999
message = b'Hello'

client_socket.sendto(message, (host_ip, port))
fps, st, frames_to_count, cnt = (0, 0, 20, 0)
while True:
    packet, _ = client_socket.recvfrom(BUFF_SIZE)
    data = base64.b64decode(packet, ' /')
    npdata = np.fromstring(data, dtype=np.uint8)
    frame = cv2.imdecode(npdata, 1)
    frame = cv2.putText(frame, 'FPS: ' + str(fps), (10, 40), cv2.FONT_HERSHEY_SIMPLEX, 0.7, (0, 0, 255), 2)
    cv2.imshow("RECEIVING VIDEO", frame)
    key = cv2.waitKey(1) & 0xFF
    if key == ord('q'):
        client_socket.close()
        break
    if cnt == frames_to_count:
        try:
            fps = round(frames_to_count / (time.time() - st))
            st = time.time()
            cnt = 0
        except:
            pass
    cnt += 1
