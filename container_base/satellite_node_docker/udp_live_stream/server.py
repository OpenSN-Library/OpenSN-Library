# This is server code to send video frames over UDP
import base64
import time
import cv2
import imutils
import socket

BUFF_SIZE = 65536
server_socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_RCVBUF, BUFF_SIZE)
server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_DEBUG, 1)
server_socket.setsockopt(socket.IPPROTO_IP, socket.IP_OPTIONS,bytearray([0x93, 0x6, 0x1, 0x2, 0x3, 0x4]))
host_name = socket.gethostname()
host_ip = '0.0.0.0'  # socket.gethostbyname(host_name)
print(host_ip)
port = 9999
socket_address = (host_ip, port)
server_socket.bind(socket_address)
print('Listening at:', socket_address)

vid = cv2.VideoCapture("movie_clip.mp4")  # replace 'rocket.mp4' with 0 for webcam
fps, st, frames_to_count, cnt = (0, 0, 20, 0)

while True:
    msg, client_addr = server_socket.recvfrom(BUFF_SIZE)
    print('GOT connection from ', client_addr)
    WIDTH = 400
    while (vid.isOpened()):
        _, frame = vid.read()
        
        frame = imutils.resize(frame, width=WIDTH)
        encoded, buffer = cv2.imencode('.jpg', frame, [cv2.IMWRITE_JPEG_QUALITY, 80])
        message = base64.b64encode(buffer)
        server_socket.sendto(message, client_addr)
        frame = cv2.putText(frame, 'FPS: ' + str(fps), (10, 40), cv2.FONT_HERSHEY_SIMPLEX, 0.7, (0, 0, 255), 2)
        cv2.imshow('TRANSMITTING VIDEO', frame)
        key = cv2.waitKey(1) & 0xFF
        if key == ord('q'):
            server_socket.close()
            break
        if cnt == frames_to_count:
            try:
                fps = round(frames_to_count / (time.time() - st))
                st = time.time()
                cnt = 0
            except:
                pass
        cnt += 1
