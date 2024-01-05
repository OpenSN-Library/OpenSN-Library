import threading
from node_list_watcher import NodeListWatcher
from position_calculator import calculate
import time



if __name__ == "__main__":
    watcher_thread = NodeListWatcher()
    calculate_thread = threading.Thread(target=calculate)
    watcher_thread.start()
    calculate_thread.start()
    watcher_thread.join()
    calculate_thread.join()