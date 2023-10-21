import os


def ban_ip(ip: str):
    cmd = 'iptables -I INPUT -s %s -j DROP' % ip
    os.system(cmd)


def allow_ip(ip: str):
    cmd = 'iptables -D INPUT -s %s -j DROP' % ip
    os.system(cmd)
