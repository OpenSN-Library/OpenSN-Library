#!/bin/sh
# 把 canonical 的 SDK（opensn/）同步到 Standard/opensn/。
# Standard 的运行方式要求 SDK 在它自己的目录里，所以这里保留一份副本。
set -e
cd "$(dirname "$0")"
rm -rf Standard/opensn
cp -R opensn Standard/opensn
find Standard/opensn -name '__pycache__' -type d -prune -exec rm -rf {} + 2>/dev/null || true
echo "synced opensn -> Standard/opensn"
