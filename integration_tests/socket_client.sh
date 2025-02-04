#!/bin/bash

SOCKET_PATH="/tmp/pipeline-control.sock"

# Check if netcat (nc) is installed
if ! command -v nc &> /dev/null; then
    echo "Error: netcat (nc) is required but not installed."
    exit 1
fi

# Check if socket exists
if [ ! -e "$SOCKET_PATH" ]; then
    echo "Error: Socket $SOCKET_PATH does not exist."
    exit 1
fi

echo "Connecting to $SOCKET_PATH..."

# Clean up on script exit
cleanup() {
    echo -e "\nCleaning up..."
    exit 0
}

# Set up trap for clean exit
trap cleanup SIGINT SIGTERM

# Use netcat to connect to the Unix socket and echo all received data
while true; do
    echo "Listening for messages..."
    nc -U "$SOCKET_PATH" | while read -r line; do
        echo "Received: $line"
        echo "$line" | nc -U "$SOCKET_PATH"
        echo "Echoed back: $line"
    done
    
    echo "Connection closed. Attempting to reconnect in 2 seconds..."
    sleep 2
done
