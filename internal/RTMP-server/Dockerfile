FROM tiangolo/nginx-rtmp

# Copy our custom nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Create HLS directory
RUN mkdir -p /tmp/hls

# Create directory for static files
RUN mkdir -p /usr/share/nginx/html

# Copy the player HTML file
COPY player.html /usr/share/nginx/html/player.html 