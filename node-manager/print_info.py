import requests


def print_info():
    port = 8080
    # 进行post请求的发送
    requests.post(url='http://127.0.0.1:%d/api/satellite/print' % port)


if __name__ == "__main__":
    print_info()
