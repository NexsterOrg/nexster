#!/bin/bash

# List of port numbers
port_numbers=(8000 8001 8002 8003 8004)

for port in "${port_numbers[@]}"; do
    # Find process ID (PID) using lsof and grep for the given port
    pid=$(lsof -t -i :$port)
    if [ -n "$pid" ]; then
        echo "Process running on port $port with PID: $pid - Terminating..."
        # Kill the process with SIGKILL (-9)
        kill -9 "$pid"
    else
        echo "No process found running on port $port"
    fi
done

echo "--done--"
