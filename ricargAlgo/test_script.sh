#!/bin/bash
# Bash script to run multiple nodes for testing Ricart-Agrawala algorithm
# Usage: ./test_script.sh

echo "Starting Ricart-Agrawala Algorithm Test"
echo "========================================"
echo ""
echo "This will start 3 nodes in separate terminal windows"
echo "Press Ctrl+C to stop all nodes"
echo ""

# Get the current directory
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Start Node A
gnome-terminal -- bash -c "cd '$DIR' && go run . --id A --listen :50051 --mode manual; exec bash" 2>/dev/null || \
xterm -e "cd '$DIR' && go run . --id A --listen :50051 --mode manual" 2>/dev/null || \
osascript -e "tell app \"Terminal\" to do script \"cd '$DIR' && go run . --id A --listen :50051 --mode manual\"" 2>/dev/null &
sleep 0.5

# Start Node B
gnome-terminal -- bash -c "cd '$DIR' && go run . --id B --listen :50052 --mode manual; exec bash" 2>/dev/null || \
xterm -e "cd '$DIR' && go run . --id B --listen :50052 --mode manual" 2>/dev/null || \
osascript -e "tell app \"Terminal\" to do script \"cd '$DIR' && go run . --id B --listen :50052 --mode manual\"" 2>/dev/null &
sleep 0.5

# Start Node C
gnome-terminal -- bash -c "cd '$DIR' && go run . --id C --listen :50053 --mode manual; exec bash" 2>/dev/null || \
xterm -e "cd '$DIR' && go run . --id C --listen :50053 --mode manual" 2>/dev/null || \
osascript -e "tell app \"Terminal\" to do script \"cd '$DIR' && go run . --id C --listen :50053 --mode manual\"" 2>/dev/null &

echo "All nodes started!"
echo "In each window, press ENTER to request critical section"
echo ""
echo "Test Scenarios:"
echo "1. Request CS from different nodes simultaneously"
echo "2. Request CS sequentially"
echo "3. Verify mutual exclusion (only one node in CS at a time)"
echo "4. Check Lamport timestamps are consistent"

