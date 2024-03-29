echo 4096 > /proc/sys/fs/inotify/max_user_instances
echo 8192 > /proc/sys/net/ipv4/neigh/default/gc_thresh1 
echo 16384 > /proc/sys/net/ipv4/neigh/default/gc_thresh2 
echo 32768 > /proc/sys/net/ipv4/neigh/default/gc_thresh3
iptables -A FORWARD -j ACCEPT