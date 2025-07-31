#!/bin/bash

# Webhook handler for RTMP events
# This script is called by nginx-rtmp when streams are published/unpublished

# Get the stream key from nginx
STREAM_KEY="$1"
ACTION="$2"
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Determine action based on the event
if [ "$ACTION" = "publish" ]; then
    EVENT_ACTION="publish"
elif [ "$ACTION" = "publish_done" ]; then
    EVENT_ACTION="unpublish"
else
    EVENT_ACTION="unknown"
fi

# Create JSON payload
JSON_PAYLOAD=$(cat <<EOF
{
  "stream_key": "$STREAM_KEY",
  "action": "$EVENT_ACTION",
  "timestamp": "$TIMESTAMP"
}
EOF
)

# Send webhook to encoder service
curl -X POST http://encoder:8082/events/published \
  -H "Content-Type: application/json" \
  -d "$JSON_PAYLOAD" \
  --connect-timeout 5 \
  --max-time 10 \
  --silent \
  --show-error

# Log the event
echo "$(date): $EVENT_ACTION event for stream: $STREAM_KEY" >> /var/log/nginx/webhook.log 