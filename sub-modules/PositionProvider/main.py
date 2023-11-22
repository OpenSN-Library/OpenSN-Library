from global_var import *
import threading
from etcd_watcher import watch_instance
from position_calculator import calculate
import time



if __name__ == "__main__":
    watcher_thread = threading.Thread(target=watch_instance)
    calculate_thread = threading.Thread(target=calculate)
    watcher_thread.start()
    calculate_thread.start()
    watcher_thread.join()
    calculate_thread.join()