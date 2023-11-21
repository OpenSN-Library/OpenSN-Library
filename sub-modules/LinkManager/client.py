'''
客户端代码

'''
import socket 

#创建套接字
tcpClientSocket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
print('socket---%s'%tcpClientSocket)
#链接服务器
serverAddr = ('192.168.77.3',8080)
tcpClientSocket.connect(serverAddr)
print('connect success!')

while True:
    #发送数据
    sendData = raw_input('please input the send message:')

    if len(sendData)>0:
        tcpClientSocket.send(sendData)  

    else:
        break   

    #接收数据
    recvData = tcpClientSocket.recv(1024)
    #打印接收到的数据
    print('the receive message is:%s'%recvData)

#关闭套接字
tcpClientSocket.close()
print('close socket!')