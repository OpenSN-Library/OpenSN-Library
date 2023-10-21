# all satellites
import threading
from collections import OrderedDict

satellites = []
networks = {}
satellite_map = {}
connect_order_map = OrderedDict()
interface_map = {}

# lock
interface_map_lock = threading.Lock()