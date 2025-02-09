#!/bin/bash

# UUID of the pipeline to cancel. Must be changed when running
PIPELINE_ID="c4548a16-7400-4291-8593-9929d0301d82"

# Generate the JSON-RPC cancellation request
JSON_PAYLOAD=$(cat <<EOF
{
  "jsonprc": "2.0",
  "id": 1,
  "method": "pipeline-cancelation",
  "params": {
    "pipeline-id": "$PIPELINE_ID",
    "pipeline-lt-secret": "$PIPELINE_SECRET"
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
if [ $? -eq 0 ]; then
  echo "Cancellation request for pipeline $PIPELINE_ID sent successfully."
else
  echo "Failed to send cancellation request. Ensure 'socat' is installed and the server is running."
  exit 1
fi
