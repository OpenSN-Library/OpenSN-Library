import threading
from node_list_watcher import NodeListWatcher
from position_calculator import evnet_generator
import time



if __name__ == "__main__":
    watcher_thread = NodeListWatcher()
    calculate_thread = threading.Thread(target=evnet_generator)
    watcher_thread.start()
    calculate_thread.start()
    watcher_thread.join()
    calculate_thread.join()