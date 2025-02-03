#!/bin/bash

# UUID of the pipeline to cancel
PIPELINE_ID="e3f0c703-3c36-4ed4-9aac-ab27b39d9f79"
# Placeholder secret (adjust if the server expects a specific value)
PIPELINE_SECRET="dummy-secret"

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
