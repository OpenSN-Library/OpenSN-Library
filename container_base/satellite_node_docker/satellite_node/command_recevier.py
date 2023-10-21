import logging
import socket
import os
import json
import time


def tcp_receiver(host_ip):
    tcp_server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    tcp_server_socket.bind((host_ip, 5000))
    while True:
        try:
            tcp_server_socket.listen(64)
            client_socket = tcp_server_socket.accept()
            recv_data = client_socket[0].recv(8192)
            req = json.loads(recv_data.decode('utf-8'))
            if req['method'] == 'traceroute':
                resp = traceroute(req['args'][0])
                resp_bytes = json.dumps(resp)
                client_socket[0].send(resp_bytes.encode('utf-8'))
            elif req['method'] == 'start_send':
                resp = start_sender(req['args'][0],req['args'][1],req['args'][2])
                resp_bytes = json.dumps(resp)
                client_socket[0].send(resp_bytes.encode('utf-8'))
            elif req['method'] == 'start_comp':
                resp = start_compresser()
                resp_bytes = json.dumps(resp)
                client_socket[0].send(resp_bytes.encode('utf-8'))
            elif req['method'] == 'start_recv':
                resp = start_recevier()
                resp_bytes = json.dumps(resp)
                client_socket[0].send(resp_bytes.encode('utf-8'))
        except Exception as e:
            logging.error("TCP Error" + str(e))


def traceroute(dst_ip: str) -> dict:
    print("exec traceroute")
    try:
        result_reader = os.popen("traceroute %s" % dst_ip)
        result = result_reader.read()
        splt = result.split('\n')
        if len(splt) > 1:
            ans = []
            for item in splt[1:-1]:
                ans.append(item.split())
            return {
                'code': 0,
                'message': 'unreachable',
                'data': ans
            }
        return {
            'code': -2,
            'message': 'unreachable',
            'data': None
        }
    except Exception as e:
        return {
            'code': -1,
            'message': str(e),
            'data': None
        }

def start_sender(compresser_ip: str, target_ip: str, another_target_ip:str) -> dict:
    print("exec start sender")
    try:
        log_file1 = "send_log1_%d.log"%int(time.time())
        log_file2 = "send_log2_%d.log"%int(time.time())
        result_reader = os.popen("nohup python3 /video_trans/start_send.py %s %s > /video_trans/%s 2>&1 &"%(target_ip,target_ip,log_file1))
        result_reader = os.popen("nohup python3 /video_trans/start_send.py %s %s > /video_trans/%s 2>&1 &"%(another_target_ip,another_target_ip,log_file2))
        return {
            'code': 0,
            'message': "success",
            'data': log_file1 + " " + log_file2
        }
    except Exception as e:
        return {
            'code': -1,
            'message': str(e),
            'data': None
        }

def start_compresser() -> dict:
    print("exec start compresser")
    try:
        log_file = "compresser_log_%d.log"%int(time.time())
        result_reader = os.popen("nohup python3 /video_trans/start_compress.py > /video_trans/%s 2>&1 &"%log_file)
        return {
            'code': 0,
            'message': "success",
            'data': log_file
        }
    except Exception as e:
        return {
            'code': -1,
            'message': str(e),
            'data': None
        }

def start_recevier() -> dict:
    print("exec start recevier")
    try:
        log_file = "recv_log_%d.log"%int(time.time())
        result_reader = os.popen("nohup python3 /video_trans/start_recv.py > /video_trans/%s 2>&1 &"%log_file)
        return {
            'code': 0,
            'message': "success",
            'data': log_file
        }
    except Exception as e:
        return {
            'code': -1,
            'message': str(e),
            'data': None
        }

