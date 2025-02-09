#!/bin/bash

JSON_PAYLOAD=$(cat <<EOF
{
    "jsonprc": "2.0",
    "id": 1,
    "method": "get-reports",
    "params": {
        "pipeline-name": "$1",
        "type": "json"
     }
}
EOF
)

# Calculate the exact byte length of the JSON payload
CONTENT_LENGTH=$(echo -n "$JSON_PAYLOAD" | wc -c)

# Construct the full message with headers
HEADER="Content-Length: $CONTENT_LENGTH\r\n\r\n"
FULL_MESSAGE="$HEADER$JSON_PAYLOAD"

# Use socat to send the message via the Unix socket
echo -ne "$FULL_MESSAGE" | socat - UNIX-CONNECT:/tmp/pipeline-control.sock

# Check for command success
if [ $? -ne 0 ]; then
  echo "Failed to send get request. Ensure 'socat' is installed and the server is running."
  exit 1
fi
