#!/bin/bash
# Start System Health Monitor
# Dashboard: http://localhost:9091
# API: http://localhost:9092/api/health
DIR="$(cd "$(dirname "$0")" && pwd)"
echo "Starting System Health Monitor..."
echo "  Dashboard: http://localhost:9091"
echo "  API:       http://localhost:9092/api/health"
echo ""
echo "Run these in separate terminals:"
echo "  python3 $DIR/api.py 9092"
echo "  python3 -m http.server 9091 --directory $DIR"
