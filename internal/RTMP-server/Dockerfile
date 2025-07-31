FROM tiangolo/nginx-rtmp

# Install curl for webhook calls
RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

# Copy our custom nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Copy webhook handler script
COPY webhook_handler.sh /usr/local/nginx/webhook_handler.sh
RUN chmod +x /usr/local/nginx/webhook_handler.sh

# Create HLS directory
RUN mkdir -p /tmp/hls

# Create directory for static files
RUN mkdir -p /usr/share/nginx/html

# Copy the player HTML file
COPY player.html /usr/share/nginx/html/player.html

# Create log directory
RUN mkdir -p /var/log/nginx 