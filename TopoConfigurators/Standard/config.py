"""运行参数（全部可用环境变量覆盖）。"""

import os

STEP_SECOND = float(os.getenv("STEP_SECOND", "5"))
POLAR_THRESHOLD_DEG = float(os.getenv("POLAR_THRESHOLD_DEG", "66.5"))
BANDWIDTH_BPS = int(os.getenv("BANDWIDTH_BPS", "8000000"))
LOSS_RATIO = float(os.getenv("LOSS_RATIO", "0.015"))
