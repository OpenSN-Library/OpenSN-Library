#coding=utf-8
#write by zxy987872674
'''
服务器端代码

'''
import socket
#创建套接字tcp
tcpServerSocket = socket.socket(socket.AF_INET,socket.SOCK_STREAM)
address = ('',8080)
tcpServerSocket.bind(address)
tcpServerSocket.listen(5)
while True:
    newServerSocket,destAddr = tcpServerSocket.accept()
    while True: 

        recvData = newServerSocket.recv(1024)
        if len(recvData)>0:
            newServerSocket.send('thanks!')
        elif len(recvData) == 0:
            newServerSocket.close()
            print('----------')
            break

tcpServerSocket.close()

