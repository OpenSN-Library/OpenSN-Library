# cd /edit && nohup python3 start_compress.py > compress_log.log 2>&1 &
trap "exit 0" TERM
cd /satellite_node/ && ./bootstrap.sh

