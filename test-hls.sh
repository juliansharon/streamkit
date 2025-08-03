#!/bin/bash

echo "=== Testing HLS Endpoints ==="

# Test health endpoint
echo "1. Testing health endpoint..."
curl -s http://localhost:8082/health | jq .

echo -e "\n2. Testing HLS playlist endpoint..."
echo "URL: http://localhost:8082/hls/941e0b13-1ba0-4152-9e57-be7727ac15b8/playlist.m3u8"
curl -I http://localhost:8082/hls/941e0b13-1ba0-4152-9e57-be7727ac15b8/playlist.m3u8

echo -e "\n3. Accessing MinIO directly..."
echo "MinIO Console: http://localhost:9001"
echo "Username: minioadmin"
echo "Password: minioadmin"

echo -e "\n4. Direct MinIO file access..."
echo "Try accessing: http://localhost:9000/hls-streams/hls/941e0b13-1ba0-4152-9e57-be7727ac15b8/playlist.m3u8"

echo -e "\n5. To view in browser:"
echo "- Open: http://localhost:9001 (MinIO Console)"
echo "- Login with minioadmin/minioadmin"
echo "- Navigate to hls-streams bucket"
echo "- Find your stream folder: hls/941e0b13-1ba0-4152-9e57-be7727ac15b8/"
echo "- Click on playlist.m3u8 to view/download"

echo -e "\n6. To test with VLC:"
echo "- Open VLC Media Player"
echo "- Media → Open Network Stream"
echo "- Enter: http://localhost:8082/hls/941e0b13-1ba0-4152-9e57-be7727ac15b8/playlist.m3u8" 